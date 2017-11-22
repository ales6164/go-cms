package api

import (
	"encoding/json"
	"reflect"
	"fmt"
)

func (e *Entity) FromForm(c Context) (*DataHolder, error) {
	var h = e.New(c)

	// todo: fix this
	c.r.FormValue("a")

	err := c.r.ParseForm()
	if len(c.r.Form) != 0 {
		for name, values := range c.r.Form {
			// remove '[]' from fieldName if it's an array
			if len(name) > 2 && name[len(name)-2:] == "[]" {
				name = name[:len(name)-2]
			}

			for _, v := range values {
				/*log.Infof(c.Context, "Appending '%s' value: %v", name, v)*/

				err = h.appendValue(name, v, Low)
				if err != nil {
					return h, err
				}
			}
		}
	} else if len(c.r.PostForm) != 0 {
		for name, values := range c.r.PostForm {
			// remove '[]' from fieldName if it's an array
			if len(name) > 2 && name[len(name)-2:] == "[]" {
				name = name[:len(name)-2]
			}

			for _, v := range values {
				err = h.appendValue(name, v, Low)
				if err != nil {
					return h, err
				}
			}
		}
	} else {
		return e.FromBody(c)
	}

	return h, err
}

func (e *Entity) FromBody(c Context) (*DataHolder, error) {
	var err error

	c = c.WithBody()

	if len(c.body.body) == 0 {
		return e.New(c), nil
	}

	var t map[string]interface{}
	err = json.Unmarshal(c.body.body, &t)
	if err != nil {
		return e.New(c), err
	}

	return e.FromMap(c, t)
}

func (e *Entity) FromMap(c Context, m map[string]interface{}) (*DataHolder, error) {
	var h = e.New(c)
	var err error

	for name, value := range m {
		if _, ok := value.([]interface{}); ok || reflect.TypeOf(value).String() == "[]interface {}" {
			for _, v := range value.([]interface{}) {
				err = h.appendValue(name, v, Base)
				if err != nil {
					return h, err
				}
			}
		} else if _, ok := value.(interface{}); ok {
			err = h.appendValue(name, value, Base)
			if err != nil {
				return h, err
			}
		} else {
			return h, fmt.Errorf(ErrFieldValueTypeNotValid, name)
		}
	}
	return h, err
}
