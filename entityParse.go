package cms

import (
	"encoding/json"
	"reflect"
	"fmt"
	"google.golang.org/appengine/datastore"
	"strings"
)

func sth() {
	data := `{"name":"Annabella", "age": 22, "hobbies": ["climbing", "playing tennis", 444], "sister": {"name": "Britney"}}`
	var t map[string]interface{}
	json.Unmarshal([]byte(data), &t)

	if _, ok := t["name"].(string); ok {
		fmt.Println("name is string")
	}

	if _, ok := t["age"].(float64); ok {
		fmt.Println("age is float64")
	}

	if _, ok := t["hobbies"].([]interface{}); ok {
		fmt.Println("hobbies is slice")
	}

	if _, ok := t["sister"].(map[string]interface{}); ok {
		fmt.Println("sister is map")
	} else {
		fmt.Println("sister is rather " + reflect.TypeOf(t["sister"]).String())
	}

	fmt.Println(t)
}

func (holder *DataHolder) PrepareFromInput() (*DataHolder, error) {
	var ctx = holder.context.WithBody()

	var m map[string]interface{}
	err := json.Unmarshal(ctx.body.body, &m)
	if err != nil {
		return holder, err
	}

	// APIRequest
	holder.request = &APIRequest{
		RemoteAddr:    ctx.r.RemoteAddr,
		Method:        ctx.r.Method,
		Host:          ctx.r.Host,
		UserAgent:     ctx.r.UserAgent(),
		Referer:       ctx.r.Referer(),
		ContentLength: ctx.r.ContentLength,
		Proto:         ctx.r.Proto,
		Body:          ctx.body.body,
	}

	if ctx.r.TLS != nil {
		holder.request.TLSVersion = ctx.r.TLS.Version
		holder.request.TLSVersion = ctx.r.TLS.CipherSuite
	}

	for _, f := range holder.Entity.Fields {

		// check for input
		if value, ok := m[f.Name]; ok {

			props, err := f.parse(holder.context, value)
			if err != nil {
				return holder, err
			}

			holder.preparedInputData[f] = props
		} else if ok, err := holder.parseNested(f, m); ok {
			// parse if input is in fieldName.childFieldName format

			if err != nil {
				return holder, err
			}
		}
	}

	return holder, nil
}

func (holder *DataHolder) parseNested(f *Field, m map[string]interface{}) (bool, error) {
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

		props, err := f.parse(holder.context, endValue)
		if err != nil {
			return true, err
		}

		holder.preparedInputData[f] = props

		return true, nil
	}

	return false, nil
}

/*func (holder *DataHolder) AppendProperty(prop datastore.Property) {
	holder.preparedInputData[prop.Name] = appendProperty(holder.preparedInputData[prop.Name], prop)
}

func (holder *DataHolder) SetProperty(prop datastore.Property) {
	holder.preparedInputData[prop.Name] = []datastore.Property{}
	holder.preparedInputData[prop.Name] = appendProperty(holder.preparedInputData[prop.Name], prop)
}*/

// appends value
func (holder *DataHolder) appendValue(dst interface{}, field *Field, value interface{}, multiple bool, lookup bool) interface{} {
	if lookup && field != nil && field.Source != nil {
		if key, ok := value.(*datastore.Key); ok {
			var dataHolder = holder.Entity.New(holder.context)
			dataHolder.key = key
			datastore.Get(holder.context.Context, key, dataHolder)
			value = dataHolder.Output(holder.context, lookup)
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
func (holder *DataHolder) appendPropertyValue(dst map[string]interface{}, prop datastore.Property, field *Field, lookup bool) map[string]interface{} {

	names := strings.Split(prop.Name, ".")
	if len(names) > 1 {
		prop.Name = strings.Join(names[1:], ".")
		if _, ok := dst[names[0]].(map[string]interface{}); !ok {
			dst[names[0]] = map[string]interface{}{}
		}
		dst[names[0]] = holder.appendPropertyValue(dst[names[0]].(map[string]interface{}), prop, field, lookup)
	} else {
		dst[names[0]] = holder.appendValue(dst[names[0]], field, prop.Value, prop.Multiple, lookup)
	}

	return dst
}

func appendProperty(field []datastore.Property, prop datastore.Property) []datastore.Property {
	return append(field, prop)
}
