package field

import (
	"fmt"
	"reflect"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
)

type Default struct {
	Name     string
	Required bool
	Multiple bool
	NoIndex  bool
	Nested   bool
}

func (x *Default) Init() error {
	return nil
}

func (x *Default) GetName() string {
	return x.Name
}

func (x *Default) GetRequired() bool {
	return x.Required
}

func (x *Default) GetMultiple() bool {
	return x.Multiple
}

func (x *Default) GetNoIndex() bool {
	return x.NoIndex
}

func (x *Default) GetNested() bool {
	return x.Nested
}

func (x *Default) Parse(value interface{}) ([]datastore.Property, error) {
	var list []datastore.Property
	if x.Multiple {
		if multiArray, ok := value.([]interface{}); ok {
			for _, value := range multiArray {
				value, err := x.Check(value)
				if err != nil {
					return list, err
				}
				list = append(list, x.Property(value))
			}
		} else if value == nil {
			value, err := x.Check(value)
			if err != nil {
				return list, err
			}
			list = append(list, x.Property(value))
		} else {
			return list, fmt.Errorf("field '%s' value type '%s' is not valid", x.Name, reflect.TypeOf(value).String())
		}
	} else {
		value, err := x.Check(value)
		if err != nil {
			return list, err
		}
		list = append(list, x.Property(value))
	}
	return list, nil
}

func (x *Default) Property(value interface{}) datastore.Property {
	return datastore.Property{
		Name:     x.Name,
		Multiple: x.Multiple,
		NoIndex:  x.NoIndex,
		Value:    value,
	}
}

func (x *Default) Check(value interface{}) (interface{}, error) {
	var err error
	if value == nil {
		if x.Required {
			return value, fmt.Errorf("field '%s' value is required", x.Name)
		}
	} else {
		err = x.Validate(value)
		if err != nil {
			return value, err
		}
		value, err = x.Transform(value)
	}
	return value, err
}

func (x *Default) Validate(value interface{}) error {
	return nil
}

func (x *Default) Transform(value interface{}) (interface{}, error) {
	return value, nil
}

func (x *Default) Output(ctx context.Context, value interface{}) interface{} {
	return value
}
