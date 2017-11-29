package api

import (
	"encoding/json"
	"reflect"
	"fmt"
)

// todo: different parsing techniques could be removed and implemented via api options - you could send some options to
// todo: the api for the parser to read from the body instead of form value ...
func (e *Entity) FromForm(c Context) (*DataHolder, error) {
	h, err := e.New(c)
	if err != nil {
		return h, err
	}

	// todo: fix this
	c.r.FormValue("a")

	err = c.r.ParseForm()
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
	h, err := e.New(c)
	if err != nil {
		return h, err
	}

	c = c.WithBody()

	if len(c.body.body) == 0 {
		return h, nil
	}

	var t map[string]interface{}
	err = json.Unmarshal(c.body.body, &t)
	if err != nil {
		return h, err
	}

	return e.FromMap(c, t)
}

func (e *Entity) FromMap(c Context, m map[string]interface{}) (*DataHolder, error) {
	h, err := e.New(c)
	if err != nil {
		return h, err
	}

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
