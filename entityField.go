package cms

import "github.com/ales6164/go-cms/api"

type Field struct {
	Label    string    `json:"label"`
	Name     string    `json:"name"`
	Type     Type      `json:"type"`
	Required bool      `json:"required"`
	Hidden   bool      `json:"hidden"`
	ReadOnly bool      `json:"readOnly"`
	Multiple bool      `json:"multiple"`
	NoIndex  bool      `json:"noIndex"`
	Rules    api.Rules `json:"rules"`

	Widget Widget `json:"widget"` // todo: based on field type widget is picked automatically; can be manually set as well

	// todo: add some group-like/embedded-entity field implementation? gob data encoding/decoding?
	Source *Entity `json:"-"` // if set, value should be encoded entity key
	//Lookup bool    `json:"lookup"` // if true, looks up entity value on output; todo: this could be query specific and not global setting

	DefaultValue interface{}        `json:"-"`
	ValueFunc    func() interface{} `json:"-"`

	ValidateFunc func(value interface{}) bool `json:"-"`
	Validate     string                       `json:"validate"`

	//GroupEntity GroupEntity `json:"groupEntity"`   // if defined, value stored as an separate entity; in field stored key
	// todo: I'd like to remove this; move evertything away from this kind?
	//Meta Meta `json:"meta"` // used for automatic admin html template creation

	// prepared functions for dealing with data
	// todo: leaving parsing to the entityParser but handling all inputs via these handlers: onInput, onValidate, on...
	// todo: these handling functions could then be assembled into one whole prepared function for the fastest response
	fieldFunc []func(ctx *ValueContext, v interface{}) (interface{}, error)
}

type Type string
type Widget string

const (
	Text           Type = "text"
	LongText       Type = "long_text"
	HTML           Type = "html"
	Date           Type = "date"
	GeoPoint       Type = "geo_point"
	Language       Type = "language"
	Time           Type = "time"
	Key            Type = "key"
	EmbeddedEntity Type = "entity"
	DateTime       Type = "date_time"
	File           Type = "file"
	Image          Type = "image"
	Number         Type = "number"
	DecimalNumber  Type = "decimal_number"
	Boolean        Type = "boolean"

	Input       Widget = "input"
	TextArea    Widget = "text_area"
	ColorPicker Widget = "color_picker"
	Select      Widget = "select"
)

// todo: move somewhere else
/*
type SearchField struct {
	Name          string
	Derived       bool
	Language      string
	TransformFunc func(value interface{}) (interface{}, error) `json:"-"`
}

type SearchFacet struct {
	Name          string
	TransformFunc func(value interface{}) (interface{}, error) `json:"-"`
}
*/
