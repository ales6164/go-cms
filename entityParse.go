package cms

import (
	"encoding/json"
	"reflect"
	"fmt"
)

func ParseBody(c Context) (map[string]interface{}, error) {
	c = c.WithBody()

	var t map[string]interface{}
	err := json.Unmarshal(c.body.body, &t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

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

// todo: parsing data
func (dh *DataHolder) Data(m map[string]interface{}) (*DataHolder, error) {

	for _, f := range dh.Entity.Fields {
		value := m[f.Name]

		if value != nil {

			if _, ok := value.([]interface{}); ok || reflect.TypeOf(value).String() == "[]interface {}" {
				for _, v := range value.([]interface{}) {
					err := h.AppendValue(name, v)
					if err != nil {
						return h, err
					}
				}
			} else if _, ok := value.(interface{}); ok {
				err := h.AppendValue(name, value)
				if err != nil {
					return h, err
				}
			} else {
				return h, fmt.Errorf(ErrFieldValueTypeNotValid, name)
			}

		} else if f.IsRequired {
			return dh, fmt.Errorf("field value '%s' required", f.Name)
		}
	}

	return h, nil
}
