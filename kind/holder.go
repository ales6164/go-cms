package kind

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"time"
	"errors"
	"encoding/json"
	"strings"
)

type Holder struct {
	Kind    *Kind `json:"entity"`
	user    *datastore.Key
	context context.Context
	key     *datastore.Key

	preparedInputData   map[*Field][]datastore.Property // user input
	hasLoadedStoredData bool
	loadedStoredData    map[string][]datastore.Property // data already stored in datastore - if exists
	savedData           []datastore.Property            // list of properties stored to datastore

	isOldVersion bool // when updating entity we want to also update old entry meta.
}

func (h *Holder) ParseInput(body []byte) error {
	var m map[string]interface{}
	err := json.Unmarshal(body, &m)
	if err != nil {
		return err
	}

	for _, f := range h.Kind.Fields {

		// check for input
		if value, ok := m[f.Name]; ok {

			props, err := f.parse(value)
			if err != nil {
				return err
			}

			h.preparedInputData[f] = props
		} else if ok, err := h.parseNested(f, m); ok {
			// parse if input is in fieldName.childFieldName format

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *Holder) parseNested(f *Field, m map[string]interface{}) (bool, error) {
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

		props, err := f.parse(endValue)
		if err != nil {
			return true, err
		}

		h.preparedInputData[f] = props

		return true, nil
	}

	return false, nil
}

// appends value
func (h *Holder) appendValue(dst interface{}, field *Field, value interface{}, multiple bool, lookup bool) interface{} {
	if lookup && field != nil && field.Source != nil {
		if key, ok := value.(*datastore.Key); ok {
			var dataHolder = h.Kind.NewHolder(h.context, h.user)
			dataHolder.key = key
			datastore.Get(h.context, key, dataHolder)
			value = dataHolder.Output(lookup)
		}
	}

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
func (h *Holder) appendPropertyValue(dst map[string]interface{}, prop datastore.Property, field *Field, lookup bool) map[string]interface{} {

	names := strings.Split(prop.Name, ".")
	if len(names) > 1 {
		prop.Name = strings.Join(names[1:], ".")
		if _, ok := dst[names[0]].(map[string]interface{}); !ok {
			dst[names[0]] = map[string]interface{}{}
		}
		dst[names[0]] = h.appendPropertyValue(dst[names[0]].(map[string]interface{}), prop, field, lookup)
	} else {
		dst[names[0]] = h.appendValue(dst[names[0]], field, prop.Value, prop.Multiple, lookup)
	}

	return dst
}

func appendProperty(field []datastore.Property, prop datastore.Property) []datastore.Property {
	return append(field, prop)
}

func (h *Holder) Output(lookup bool) map[string]interface{} {
	var output = map[string]interface{}{}

	// range over data. Value can be single value or if the field it Multiple then it's an array
	for _, prop := range h.savedData {
		output = h.appendPropertyValue(output, prop, h.Kind.fields[prop.Name], lookup)
	}

	output["id"] = h.key.Encode()

	return output
}

func (h *Holder) Load(ps []datastore.Property) error {
	h.hasLoadedStoredData = true
	h.savedData = ps
	for _, prop := range ps {
		h.loadedStoredData[prop.Name] = appendProperty(h.loadedStoredData[prop.Name], prop)
	}
	return nil
}

func (h *Holder) Save() ([]datastore.Property, error) {
	var ps []datastore.Property

	h.savedData = []datastore.Property{}

	// check if required fields are there
	for _, field := range h.Kind.Fields {
		var inputProperties = h.preparedInputData[field]
		var loadedProperties = h.loadedStoredData[field.Name]

		var toSaveProps []datastore.Property

		if len(inputProperties) != 0 {
			toSaveProps = append(toSaveProps, inputProperties...)
		} else if len(loadedProperties) != 0 {
			toSaveProps = append(toSaveProps, loadedProperties...)
		} else if field.IsRequired {
			return nil, errors.New("field " + field.Name + " required")
		}

		h.savedData = append(h.savedData, toSaveProps...)

		// save search field
		// TODO:
		/*if field.toSearchFieldConvertFunc != nil {
			if err := field.toSearchFieldConvertFunc(holder, toSaveProps...); err != nil {
				return nil, err
			}
		}*/
	}

	// set meta tags
	var now = time.Now()
	h.savedData = append(h.savedData, datastore.Property{
		Name:  "meta.updatedAt",
		Value: now,
	})
	h.savedData = append(h.savedData, datastore.Property{
		Name:  "meta.status",
		Value: "active",
	})
	if h.hasLoadedStoredData {
		if metaCreatedAt, ok := h.loadedStoredData["meta.createdAt"]; ok {
			h.savedData = append(h.savedData, metaCreatedAt[0])
		}
		if metaCreatedBy, ok := h.loadedStoredData["meta.createdBy"]; ok {
			h.savedData = append(h.savedData, metaCreatedBy[0])
		}
		if metaVersion, ok := h.loadedStoredData["meta.version"]; ok {
			var version = metaVersion[0]
			version.Value = version.Value.(int64) + 1
			h.savedData = append(h.savedData, version)
		}
	} else {
		h.savedData = append(h.savedData, datastore.Property{
			Name:  "meta.createdAt",
			Value: now,
		})
		h.savedData = append(h.savedData, datastore.Property{
			Name:  "meta.createdBy",
			Value: h.user,
		})
		h.savedData = append(h.savedData, datastore.Property{
			Name:  "meta.version",
			Value: int64(0),
		})
	}

	h.savedData = append(h.savedData, datastore.Property{
		Name:  "meta.updatedBy",
		Value: h.user,
	})

	ps = h.savedData

	return ps, nil
}

type HolderOld struct {
	data *Holder
	key  *datastore.Key
}

func (h *Holder) OldHolder(replacementKey *datastore.Key) *HolderOld {
	var old = &HolderOld{
		data: h,
		key:  replacementKey,
	}
	return old
}

func (h *HolderOld) Load(ps []datastore.Property) error {
	return nil
}

func (h *HolderOld) Save() ([]datastore.Property, error) {
	var ps []datastore.Property

	h.data.loadedStoredData["meta.status"] = []datastore.Property{{
		Name:  "meta.status",
		Value: nil,
	}}

	for _, props := range h.data.loadedStoredData {
		ps = append(ps, props...)
	}

	return ps, nil
}
