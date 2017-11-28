package api

type Field struct {
	Label    string `json:"label"`
	Name     string `json:"name"`

	Required bool   `json:"required"`
	Hidden   bool   `json:"hidden"`
	Multiple bool   `json:"multiple"`
	NoIndex  bool   `json:"noIndex"`

	Rules Rules `json:"rules"`

	// Entity -> could be type Key; Embedded Entity -> type Embedded
	// todo: type could be interface for it to be simple to embed entity or embedded entity; could also remove Entity property
	Type Type `json:"type"`

	// todo: add some group-like/embedded-entity field implementation? gob data encoding/decoding?
	Entity *Entity `json:"-"`      // if set, value should be encoded entity key
	Lookup bool    `json:"lookup"` // if true, looks up entity value on output; todo: this could be query specific and not global setting

	// todo: single Value property of type interface -> checks types on init; could have struct types for ValueFunc, ContextFunc, ...
	DefaultValue interface{}                   `json:"defaultValue"`
	ValueFunc    func() interface{}            `json:"-"`
	ContextFunc  func(ctx Context) interface{} `json:"-"`

	ValidateRgx   string                                                          `json:"validate"`
	TransformFunc func(ctx *ValueContext, value interface{}) (interface{}, error) `json:"-"`
	Validator     func(value interface{}) bool                                    `json:"-"`

	//GroupEntity GroupEntity `json:"groupEntity"`   // if defined, value stored as an separate entity; in field stored key
	// todo: I'd like to remove this; move evertything away from this kind?
	Meta Meta `json:"meta"` // used for automatic admin html template creation

	fieldFunc []func(ctx *ValueContext, v interface{}) (interface{}, error) `json:"-"`
}

type Type string
type Meta map[string]interface{}

const (
	Text          Type = "text"
	LongText      Type = "long_text"
	HTML          Type = "html"
	Date          Type = "date"
	GeoPoint      Type = "geo_point"
	Tag           Type = "tag"
	Language      Type = "language"
	Time          Type = "time"
	Key           Type = "key"
	DateTime      Type = "date_time"
	HexColor      Type = "hex_color"
	FileURL       Type = "file_url"
	ImageURL      Type = "image_url"
	Number        Type = "number"
	DecimalNumber Type = "decimal_number"
	Boolean       Type = "boolean"
)

type ValueContext struct {
	Scope Scope
	Trust ValueTrust
	Field *Field
}

type ValueTrust string

const (
	Low  ValueTrust = "low"
	Base ValueTrust = "base"
	High ValueTrust = "high"
)

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
