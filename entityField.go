package cms

import (
	"google.golang.org/appengine/datastore"
	"fmt"
	"reflect"
	"github.com/asaskevich/govalidator"
)

type Field struct {
	Label      string `json:"label"`
	Name       string `json:"name"`
	Type       Type   `json:"type"`
	IsRequired bool   `json:"isRequired"`
	/*	IsReadOnly   bool   `json:"isReadOnly"`*/
	IsNameProvider bool  `json:"isIDProvider"`
	Multiple       bool  `json:"multiple"`
	NoIndex        bool  `json:"noIndex"`
	Rules          Rules `json:"rules"`

	isNesting bool // field name is of pattern one.two

	Widget Widget `json:"widget"` // todo: based on field type widget is picked automatically; can be manually set as well

	// todo: add some group-like/embedded-entity field implementation? gob data encoding/decoding?
	Source *Entity `json:"-"` // if set, value should be encoded entity key
	//Lookup bool    `json:"lookup"` // if true, looks up entity value on output; todo: this could be query specific and not global setting

	// not yet implemented:
	DefaultValue interface{}        `json:"-"`
	ValueFunc    func() interface{} `json:"-"`

	Validate     string                       `json:"validate"` // regex pattern; only works for string values
	ValidateFunc func(value interface{}) bool `json:"-"`

	TransformFunc func(value interface{}) interface{} `json:"-"`

	// prepared functions for dealing with data
	// todo: leaving parsing to the entityParser but handling all inputs via these handlers: onInput, onValidate, on...
	// todo: these handling functions could then be assembled into one whole prepared function for the fastest response
	//fieldFunc []func(ctx *ValueContext, v interface{}) (interface{}, error)
}

// todo: make this as a prepared function as Fields property
func (f *Field) parse(ctx Context, input interface{}) ([]datastore.Property, error) {
	var list []datastore.Property

	if f.Rules != nil {
		// if rule is set, check if users rank is sufficient
		if role, ok := f.Rules[ctx.Scope]; ok && ctx.Rank < Ranks[role] {
			// users rank is lower - action forbidden
			return list, ErrForbidden
		}
	}

	// Multiple
	if f.Multiple {
		if multiArray, ok := input.([]interface{}); ok {
			for _, value := range multiArray {
				parsedValue, err := f.parseSingleValue(ctx, value)
				if err != nil {
					return list, err
				}
				list = append(list, parsedValue)
			}
		} else if input == nil {
			emptyValue, err := f.parseSingleValue(ctx, input)
			if err != nil {
				return list, err
			}
			list = append(list, emptyValue)
		} else {
			// input type not valid
			return list, fmt.Errorf("field '%s' value type '%s' is not valid", f.Name, reflect.TypeOf(input).String())
		}
	} else {
		parsedValue, err := f.parseSingleValue(ctx, input)
		if err != nil {
			return list, err
		}
		list = append(list, parsedValue)
	}

	// value func - this is left out till the end... only gets fired if field datastore property doesn't exist

	return list, nil
}

// TODO: Implement Entity as value functionality
func (f *Field) parseSingleValue(ctx Context, input interface{}) (datastore.Property, error) {
	var p datastore.Property

	if input == nil && f.IsRequired {
		return p, fmt.Errorf("field '%s' value is required", f.Name)
	}

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

	return datastore.Property{
		Name:     f.Name,
		Multiple: f.Multiple,
		Value:    input,
		NoIndex:  f.NoIndex,
	}, nil
}

type Type string
type Widget string

const (
	ID            Type = "id"
	Text          Type = "text"
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

	Input       Widget = "input"
	TextArea    Widget = "textArea"
	ColorPicker Widget = "colorPicker"
	Select      Widget = "select"
)
