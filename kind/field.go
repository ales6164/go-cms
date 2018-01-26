package kind

import (
	"fmt"
	"reflect"
	"github.com/asaskevich/govalidator"
	"google.golang.org/appengine/datastore"
)

type Field struct {
	Name       string `json:"name"`
	Type       Type   `json:"type"`
	IsRequired bool   `json:"required"`
	Multiple   bool   `json:"multiple"`
	NoIndex    bool   `json:"noIndex"`

	isNesting bool

	Source *Kind `json:"-"`

	Validate     string                       `json:"validate"` // regex pattern; only works for string values
	ValidateFunc func(value interface{}) bool `json:"-"`

	TransformFunc func(value interface{}) interface{} `json:"-"`
}

// todo: make this as a prepared function as Fields property
func (f *Field) parse(input interface{}) ([]datastore.Property, error) {
	var list []datastore.Property

	// Multiple
	if f.Multiple {
		if multiArray, ok := input.([]interface{}); ok {
			for _, value := range multiArray {
				parsedValue, err := f.parseSingleValue(value)
				if err != nil {
					return list, err
				}
				list = append(list, parsedValue)
			}
		} else if input == nil {
			emptyValue, err := f.parseSingleValue(input)
			if err != nil {
				return list, err
			}
			list = append(list, emptyValue)
		} else {
			// input type not valid
			return list, fmt.Errorf("field '%s' value type '%s' is not valid", f.Name, reflect.TypeOf(input).String())
		}
	} else {
		parsedValue, err := f.parseSingleValue(input)
		if err != nil {
			return list, err
		}
		list = append(list, parsedValue)
	}

	// value func - this is left out till the end... only gets fired if field datastore property doesn't exist

	return list, nil
}

// TODO: Implement Entity as value functionality
func (f *Field) parseSingleValue(input interface{}) (datastore.Property, error) {
	var p datastore.Property

	if input == nil {
		if f.IsRequired {
			return p, fmt.Errorf("field '%s' value is required", f.Name)
		}
	} else {
		if len(f.Validate) > 0 {
			if stringValue, ok := input.(string); ok && !govalidator.Matches(stringValue, f.Validate) {
				return p, fmt.Errorf("field '%s' value is not valid", f.Name)
			}
		} else if f.ValidateFunc != nil {
			if !f.ValidateFunc(input) {
				return p, fmt.Errorf("field '%s' value is not valid", f.Name)
			}
		}
		if f.TransformFunc != nil {
			input = f.TransformFunc(input)
			if err, ok := input.(error); ok {
				return p, err
			}
		}
	}

	return datastore.Property{
		Name:     f.Name,
		Multiple: f.Multiple,
		Value:    input,
		NoIndex:  f.NoIndex,
	}, nil
}

type Type string

const (
	ID            Type = "id"
	Category      Type = "category"
	Text          Type = "text"
	Slug          Type = "slug"
	LongText      Type = "longText"
	Timestamp     Type = "timestamp"
	HTML          Type = "html"
	GeoPoint      Type = "geoPoint"
	Language      Type = "language"
	Key           Type = "key"
	NestedEntity  Type = "nestedEntity"
	Number        Type = "number"
	URL           Type = "url"
	DecimalNumber Type = "decimalNumber"
	Boolean       Type = "boolean"
)
