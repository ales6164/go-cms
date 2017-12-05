package cms

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/appengine/datastore"
	"strings"
	"time"
)

// PreparedEntity data holder
type DataHolder struct {
	context Context
	Entity  *Entity `json:"-"`

	isNew             bool
	keepExistingValue bool // turn this true when receiving old data from database; used for editing existing entity
	lastScope         Scope // to know how to treat new data; if needs checking

	id    string                 // saved during datastore operations and returned on output
	data  []datastore.Property                   // this can be edited by load/save, and conditionally with appendField functions
}

const (
	ErrNamedFieldNotDefined                string = "named field '%s' is not defined"
	ErrDatastoreFieldPropertyMultiDismatch string = "datastore field '%s' doesn't match in property multi"
	ErrFieldRequired                       string = "field '%s' required"
	ErrFieldEditPermissionDenied           string = "field '%s' edit permission denied"
	ErrFieldValueNotValid                  string = "field '%s' value is not valid"
	ErrFieldValueTypeNotValid              string = "field '%s' value type is not valid"
	ErrValueIsNil                          string = "field '%s' value is empty"
)

func init() {
	gob.Register(time.Now())
}

func (h *DataHolder) Get(ctx Context, name string) interface{} {
	var endValue interface{}

	sep := strings.Split(name, ".")
	if field, ok := h.Entity.fields[sep[0]]; ok {
		var value = h.data[field]

		if field.Lookup && field.Entity != nil {
			if field.Multiple {
				endValue = []interface{}{}

				for _, v := range value.([]interface{}) {
					endV, _ := field.Entity.Lookup(ctx, v.(string))
					endValue = append(endValue.([]interface{}), endV)
				}
			} else {
				endValue, _ = field.Entity.Lookup(ctx, value.(string))
			}
		} else {
			endValue = value
		}

		if len(sep) > 1 {
			for i := 1; i < len(sep); i++ {
				if endMap, ok := endValue.(map[string]interface{}); ok {
					endValue = endMap[sep[i]]
					/*if endValue, ok = endMap[sep[i]]; ok {}*/
				} else {
					return nil
				}
			}
		}

		return endValue
	}
	return nil
}

func output(ctx Context, id string, data Data, cacheLookup bool) map[string]interface{} {
	var output = map[string]interface{}{}

	// range over data. Value can be single value or if the field it Multiple then it's an array
	for field, value := range data {
		if field.Hidden {
			continue
		}

		if cacheLookup && field.Lookup && field.Entity != nil {
			if field.Multiple {
				for _, v := range value.([]interface{}) {
					v, _ = field.Entity.Lookup(ctx, v.(string))
					output[field.Name] = appendValue(output[field.Name], v, true)
				}
			} else {
				value, _ = field.Entity.Lookup(ctx, value.(string))
				output[field.Name] = appendValue(output[field.Name], value, false)
			}
		} else {
			output[field.Name] = appendValue(output[field.Name], value, false)
		}
	}

	output["id"] = id

	return output
}

func appendValue(field interface{}, value interface{}, multiple bool) interface{} {
	if multiple {
		if field == nil {
			field = []interface{}{}
		}
		field = append(field.([]interface{}), value)
	} else {
		field = value
	}
	return field
}

func flatOutput(id string, data Data) map[string]interface{} {
	var output = map[string]interface{}{}

	for field, value := range data {
		if field.Hidden {
			continue
		}

		output[field.Name] = value
	}

	output["id"] = id

	return output
}

func (h *DataHolder) Output(ctx Context) map[string]interface{} {
	return output(ctx, h.id, h.data, true)
}

func (h *DataHolder) FlatOutput() map[string]interface{} {
	return flatOutput(h.id, h.data)
}

func (h *DataHolder) JSON(ctx Context) (string, error) {
	bs, err := json.Marshal(h.Output(ctx))
	return string(bs), err
}

// Appends any kind of value to the named property
// Skips with no errors if field doesn't exist
func (dh *DataHolder) AppendValue(name string, value interface{}) error {
	if field, ok := dh.Entity.fields[name]; ok {
		return dh.appendFieldValue(field, value)
	}
	return nil
}

// Safely appends value
func (e *DataHolder) appendFieldValue(field *Field, value interface{}) error {
	var v = value
	var err error
	for _, fun := range field.fieldFunc {
		v, err = fun(vc, v)
		if err != nil {
			return err
		}
	}

	if v != nil {
		return e.unsafeAppendFieldValue(field, v, value, vc.Scope, e.keepExistingValue)
	}

	return fmt.Errorf(ErrValueIsNil, field.Name)
}

// UNSAFE Appends value without any checks
func (h *DataHolder) unsafeAppendFieldValue(field *Field, value interface{}, formValue interface{}, scope Scope, keepExistingValue bool) error {
	if role, ok := field.Rules[scope]; ok {
		if h.context.Rank < Ranks[role] {
			return ErrNotAuthorized
		}
	}

	if field.Multiple {
		// Todo: Check if this check is necessary
		if _, ok := h.data[field]; !ok {
			h.data[field] = []interface{}{}
		} else if keepExistingValue {
			return nil
		}
		if _, ok := h.data[field].([]interface{}); !ok {
			panic(errors.New("field '" + field.Name + "' value is not []interface{}"))
		}
		h.data[field] = append(h.data[field].([]interface{}), value)
	} else {
		if _, ok := h.data[field]; ok && keepExistingValue {
			return nil
		}
		h.data[field] = value
	}
	return nil
}

// load from datastore properties into Data map
func (e *DataHolder) Load(ps []datastore.Property) error {
	/*e.data = map[*Field]interface{}{}*/
	for _, prop := range ps {
		if field, ok := e.Entity.fields[prop.Name]; ok {
			if prop.Multiple != field.Multiple {
				return fmt.Errorf(ErrDatastoreFieldPropertyMultiDismatch, prop.Name)
			}
			e.unsafeAppendFieldValue(field, prop.Value, nil, Read, e.keepExistingValue)
		} else {
			return fmt.Errorf(ErrNamedFieldNotDefined, prop.Name)
		}
	}
	return nil
}

// load Data map into datastore Property array
func (e *DataHolder) Save() ([]datastore.Property, error) {
	var ps []datastore.Property

	// check if required fields are there
	for _, field := range e.Entity.requiredFields {
		if _, ok := e.data[field]; !ok {
			return ps, fmt.Errorf(ErrFieldRequired, field.Name)
		}
	}

	// create datastore property list
	for field, value := range e.data {
		// set group name

		datastore.Entity{}

		if field.Multiple {
			for _, v := range value.([]interface{}) {
				ps = append(ps, datastore.Property{
					Name:     field.Name,
					Multiple: field.Multiple,
					Value:    v,
					NoIndex:  field.NoIndex,
				})
			}
		} else {
			ps = append(ps, datastore.Property{
				Name:     field.Name,
				Multiple: field.Multiple,
				Value:    value,
				NoIndex:  field.NoIndex,
			})
		}
	}

	return ps, nil
}
