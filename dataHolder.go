package cms

import (
	"fmt"
	"google.golang.org/appengine/datastore"
	"time"
)

// PreparedEntity data holder
type DataHolder struct {
	context Context
	Entity  *Entity `json:"entity"`

	/*	isNew             bool
		keepExistingValue bool  // turn this true when receiving old data from database; used for editing existing entity
		lastScope         Scope // to know how to treat new data; if needs checking*/

	key *datastore.Key

	waiting map[*Field]bool // after first parse we can put empty fields in here for the second parse
	Data    []datastore.Property `json:"data"`
	Meta    Meta                 `json:"meta"`

	nameProviderValue interface{} // value for entity.NameFunc
}

type Meta struct {
	Name      string         `datastore:"name",json:"name"`
	CreatedAt time.Time      `datastore:"createdAt",json:"createdAt"`
	UpdatedAt time.Time      `datastore:"updatedAt",json:"updatedAt"`
	CreatedBy *datastore.Key `datastore:"createdBy",json:"createdBy"`
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

func (e *Entity) New(ctx Context) *DataHolder {
	var dataHolder = &DataHolder{
		Entity:  e,
		context: ctx,
		waiting: map[*Field]bool{},
		Meta:    Meta{},
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

			h.Data = append(h.Data, props...)
		} else {
			// map doesn't contain this field; this could mean we don't want to change exiting value OR the field value
			// is not required; required checks are done in on Save() function
			h.waiting[f] = true
		}
	}

	return h, err
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

func (h *DataHolder) Output(ctx Context) map[string]interface{} {
	var output = map[string]interface{}{}

	// range over data. Value can be single value or if the field it Multiple then it's an array
	for _, property := range h.Data {
		if f, ok := h.Entity.fields[property.Name]; ok {
			if f.Rules != nil {
				// if rule is set, check if users rank is sufficient
				if role, ok := f.Rules[ctx.Scope]; ok && ctx.Rank < Ranks[role] {
					// users rank is lower - action forbidden
					// skip this field
					continue
				}
			}
			output[f.Name] = appendValue(output[f.Name], property.Value, f.Multiple)
		}
	}

	output["id"] = h.key.Encode()
	output["meta"] = h.Meta

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
	if len(h.Data) > 0 {

		for _, property := range ps {
			if property.Name == "meta" {
				h.Meta = property.Value.(Meta)
				continue
			}

			if f, ok := h.Entity.fields[property.Name]; ok && h.waiting[f] {

				newProperty, err := f.parseSingleValue(h.context, property.Value)
				if err != nil {
					return fmt.Errorf("error loading stored entity as field '%s' value breaks new parsing requirements: %s", f.Name, err.Error())
				}

				h.Data = append(h.Data, newProperty)
			}
		}

	} else {
		for i, property := range ps {
			if property.Name == "meta" {
				h.Meta = property.Value.(Meta)

				ps = remove(ps, i)
				break
			}
		}

		h.Data = ps
	}

	return nil
}

func (h *DataHolder) Save() ([]datastore.Property, error) {
	// check if required fields are there
	for f, isEmpty := range h.waiting {
		if f.IsRequired && isEmpty {
			return h.Data, fmt.Errorf("error saving data: field '%s' value is required", f.Name)
		}
	}

	return append([]datastore.Property{{
		Value:    h.Meta,
		Name:     "meta",
		Multiple: false,
		NoIndex:  false,
	}}, h.Data...), nil
}

func remove(s []datastore.Property, i int) []datastore.Property {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}