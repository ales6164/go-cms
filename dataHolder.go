package cms

import (
	"fmt"
	"google.golang.org/appengine/datastore"
	"strings"
)

// PreparedEntity data holder
type DataHolder struct {
	context Context
	Entity  *Entity `json:"entity"`

	/*	isNew             bool
		keepExistingValue bool  // turn this true when receiving old data from database; used for editing existing entity
		lastScope         Scope // to know how to treat new data; if needs checking*/

	key *datastore.Key

	prepared bool              // true if has prepared data
	writer   map[*Field]wrtOpt // if updating entity we want to push updated fields here to keep them from rewriting on second load
	Data     map[string][]datastore.Property `json:"data"`

	nameProviderValue interface{} // value for entity.NameFunc
}

type wrtOpt int

const (
	PreventRewrite wrtOpt = 0
	Rewritable     wrtOpt = 1
	Required       wrtOpt = 2
)

const (
	ErrNamedFieldNotDefined                string = "named field '%s' is not defined"
	ErrDatastoreFieldPropertyMultiDismatch string = "datastore field '%s' doesn't match in property multi"
	ErrFieldRequired                       string = "field '%s' required"
	ErrFieldEditPermissionDenied           string = "field '%s' edit permission denied"
	ErrFieldValueNotValid                  string = "field '%s' value is not valid"
	ErrFieldValueTypeNotValid              string = "field '%s' value type is not valid"
	ErrValueIsNil                          string = "field '%s' value is empty"
)

func (e *Entity) New(ctx Context) *DataHolder {
	var dataHolder = &DataHolder{
		Entity:  e,
		context: ctx,
		writer:  map[*Field]wrtOpt{},
		Data:    map[string][]datastore.Property{},
	}
	return dataHolder
}

// Prepares property list for datastore actions.
// If preparing multiple times on the same dataHolder,
func (h *DataHolder) Prepare(m map[string]interface{}) (*DataHolder, error) {
	var err error

	for _, f := range h.Entity.Fields {

		// map contains field
		if value, ok := m[f.Name]; ok {

			props, err := f.parse(h.context, value)
			if err != nil {
				return h, err
			}

			if f.IsNameProvider {
				h.nameProviderValue = value
			}

			h.Data[f.Name] = props
			h.writer[f] = PreventRewrite
		} else if ok, err := h.parseNested(f, m); ok {
			if err != nil {
				return h, err
			}
		} else if f.IsRequired {
			h.writer[f] = Required
		} else {
			h.writer[f] = Rewritable
		}
	}

	h.prepared = true

	return h, err
}

func (h *DataHolder) parseNested(f *Field, m map[string]interface{}) (bool, error) {
	names := strings.Split(f.Name, ".")

	if _, ok := m[names[0]]; ok {

		var endValue = m[names[0]]
		for i := 1; i < len(names); i++ {
			if nestedMap, ok := endValue.(map[string]interface{}); ok {
				endValue = nestedMap[names[i]]
			} else {
				return false, nil
			}
		}

		props, err := f.parse(h.context, endValue)
		if err != nil {
			return true, err
		}

		if f.IsNameProvider {
			h.nameProviderValue = endValue
		}

		h.Data[f.Name] = props
		h.writer[f] = PreventRewrite

		return true, nil
	}

	return false, nil
}


func (h *DataHolder) AppendProperty(prop datastore.Property) {
	h.Data[prop.Name] = appendProperty(h.Data[prop.Name], prop)
}

func (h *DataHolder) SetProperty(prop datastore.Property) {
	h.Data[prop.Name] = []datastore.Property{}
	h.Data[prop.Name] = appendProperty(h.Data[prop.Name], prop)
}

// appends value
func appendValue(dst interface{}, value interface{}, multiple bool) interface{} {
	if multiple {
		if dst == nil {
			dst = []interface{}{}
		}
		dst = append(dst.([]interface{}), value)
	} else {
		dst = value
	}
	return dst
}

// appends property to dst; it can return a flat object or structured
func appendPropertyValue(dst map[string]interface{}, prop datastore.Property) map[string]interface{} {

	names := strings.Split(prop.Name, ".")
	if len(names) > 1 {
		prop.Name = strings.Join(names[1:], ".")
		if _, ok := dst[names[0]].(map[string]interface{}); !ok {
			dst[names[0]] = map[string]interface{}{}
		}
		dst[names[0]] = appendPropertyValue(dst[names[0]].(map[string]interface{}), prop)
	} else {
		dst[names[0]] = appendValue(dst[names[0]], prop.Value, prop.Multiple)
	}

	return dst

}

func appendProperty(field []datastore.Property, prop datastore.Property) []datastore.Property {
	return append(field, prop)
}

func (h *DataHolder) Output(ctx Context) map[string]interface{} {
	var output = map[string]interface{}{}

	// range over data. Value can be single value or if the field it Multiple then it's an array
	for name, propertyList := range h.Data {
		if f, ok := h.Entity.fields[name]; ok {
			if f.Rules != nil {
				// if rule is set, check if users rank is sufficient
				if role, ok := f.Rules[ctx.Scope]; ok && ctx.Rank < Ranks[role] {
					// users rank is lower - action forbidden
					// skip this field
					continue
				}
			}
		}

		for _, prop := range propertyList {
			output = appendPropertyValue(output, prop)
		}
	}

	output["id"] = h.key.Encode()
	//output["meta"] = h.Meta

	return output
}

/*func (h *DataHolder) Get(ctx Context, name string) interface{} {
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
					*//*if endValue, ok = endMap[sep[i]]; ok {}*//*
				} else {
					return nil
				}
			}
		}

		return endValue
	}
	return nil
}*/

/*
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
}*/

// PropertyLoadSaver interface implementation. If dataHolder already contains data, it replaces only waiting fields
// Extracts Meta info
func (h *DataHolder) Load(ps []datastore.Property) error {
	if h.prepared {
		for _, prop := range ps {
			if f, ok := h.Entity.fields[prop.Name]; ok {

				if opt, ok := h.writer[f]; !ok || opt != PreventRewrite {

					newProperty, err := f.parseSingleValue(h.context, prop.Value)
					if err != nil {
						return fmt.Errorf("error loading stored entity as field '%s' value breaks new parsing requirements: %s", f.Name, err.Error())
					}

					h.Data[f.Name] = appendProperty(h.Data[f.Name], newProperty)

					h.writer[f] = Rewritable
				}
			} else {
				h.Data[prop.Name] = appendProperty(h.Data[prop.Name], prop)
			}
		}
	} else {
		for _, prop := range ps {
			h.Data[prop.Name] = appendProperty(h.Data[prop.Name], prop)
		}
	}

	return nil
}

func (h *DataHolder) Save() ([]datastore.Property, error) {
	// check if required fields are there
	for f, opt := range h.writer {
		if opt == Required {
			return nil, fmt.Errorf("error saving data: field '%s' value is required", f.Name)
		}
	}

	var props []datastore.Property

	for _, prop := range h.Data {
		props = append(props, prop...)
	}

	return props, nil
}

func remove(s []datastore.Property, i int) []datastore.Property {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}
