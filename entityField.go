package cms

import "time"

type Field struct {
	Label        string `json:"label"`
	Name         string `json:"name"`
	Type         Type   `json:"type"`
	IsRequired   bool   `json:"isRequired"`
	IsReadOnly   bool   `json:"isReadOnly"`
	IsIDProvider bool   `json:"isIDProvider"`
	Multiple     bool   `json:"multiple"`
	NoIndex      bool   `json:"noIndex"`
	Rules        Rules  `json:"rules"`
	isMeta       bool

	Widget Widget `json:"widget"` // todo: based on field type widget is picked automatically; can be manually set as well

	// todo: add some group-like/embedded-entity field implementation? gob data encoding/decoding?
	Source *Entity `json:"-"` // if set, value should be encoded entity key
	//Lookup bool    `json:"lookup"` // if true, looks up entity value on output; todo: this could be query specific and not global setting

	DefaultValue interface{}        `json:"-"`
	ValueFunc    func() interface{} `json:"-"`

	Validate     string                       `json:"validate"` // regex
	ValidateFunc func(value interface{}) bool `json:"-"`

	TransformFunc func(value interface{}) interface{} `json:"-"`

	propertyFunc func(ctx Context, formInput []string) interface{}
	// prepared functions for dealing with data
	// todo: leaving parsing to the entityParser but handling all inputs via these handlers: onInput, onValidate, on...
	// todo: these handling functions could then be assembled into one whole prepared function for the fastest response
	//fieldFunc []func(ctx *ValueContext, v interface{}) (interface{}, error)
}

func (f *Field) init() *Field {
	var fun func(ctx Context, formInput []string) interface{}
	switch f.Type {

	default:
		fun = FormOneValue
	}

	f.inputParseFunc = fun
	return f
}

func FormOneValue(ctx Context, formInput []string) interface{} {
	return formInput[0]
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

var createdAt = &Field{
	Name:    "_createdAt",
	isMeta:  true,
	NoIndex: true,
	Type:    Timestamp,
	ValueFunc: func() interface{} {
		return time.Now().UTC()
	},
}

var updatedAt = &Field{
	Name:    "_updatedAt",
	isMeta:  true,
	NoIndex: true,
	Type:    Timestamp,
	ValueFunc: func() interface{} {
		return time.Now().UTC()
	},
}

var publishedAt = &Field{
	Name:    "_publishedAt",
	isMeta:  true,
	NoIndex: true,
	Type:    Timestamp,
	/*TransformFunc: func(value interface{}) (interface{}, error) {
		var t time.Time
		if val, ok := value.(int64); ok {
			t = time.Unix(val, 0)
		} else if val, ok := value.(string); ok {
			val, err := strconv.Atoi(val)
			if err != nil {
				return t, err
			}
			t = time.Unix(int64(val), 0)
		}
		return t.UTC(), nil
	},*/
}

var createdBy = &Field{
	Name:    "_createdBy",
	isMeta:  true,
	NoIndex: true,
	Type:    Key,
	/*Entity:     User,*/
	/*ContextFunc: func(ctx Context) interface{} {
		if len(ctx.User) > 0 {
			if key, err := datastore.DecodeKey(ctx.User); err == nil {
				return key
			}
			return nil
		}
		return nil
	},*/
}

var updatedBy = &Field{
	Name:    "_updatedBy",
	isMeta:  true,
	NoIndex: true,
	Type:    Key,
	/*Entity:     User,*/
	/*ContextFunc: func(ctx Context) interface{} {
		if len(ctx.User) > 0 {
			if key, err := datastore.DecodeKey(ctx.User); err == nil {
				return key
			}
			return nil
		}
		return nil
	},*/
}
