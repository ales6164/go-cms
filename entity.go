package cms

import (
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/log"
	"regexp"
	"time"
	"strconv"
)

type Entity struct {
	Label     string `json:"label"`     // Only a-Z characters allowed
	Name      string `json:"name"`      // Only a-Z characters allowed
	Private   bool   `json:"private"`   // Protects entity with user field - only creator has access
	Protected bool   `json:"protected"` // Protects entity with password
	Cache     bool   `json:"cache"`     // Keeps values in memcache - good for categories, translations, ...

	fields map[string]*Field
	Fields []*Field `json:"fields"`

	// Called on every entity update. If url already exists (and is not the same as the previous url), calls again with failedCount increased by 1
	// todo: have a separate package for this service instead; with client id and client secret input as well
	//URLFunc func(provider *URLProvider, failedCount int) string `json:"-"`

	// Admin configuration
	Meta Meta `json:"meta"`

	preparedData map[*Field]func(ctx Context, f *Field) interface{}

	requiredFields []*Field

	indexes map[string]*DocumentDefinition

	// Rules
	Rules Rules `json:"rules"`

	// Listener
	OnInit        func(c Context, h *DataHolder) error `json:"-"`
	OnBeforeWrite func(c Context, h *DataHolder) error `json:"-"`
	OnAfterRead   func(c Context, h *DataHolder) error `json:"-"`
	OnAfterWrite  func(c Context, h *DataHolder) error `json:"-"`
}

type Parser struct {
	Field     *Field
	ParseFunc func(ctx Context, fieldName string) (interface{}, error)
}

func (e *Entity) init() (*Entity, error) {
	e.preparedData = map[*Field]func(ctx Context, f *Field) interface{}{}

	for _, field := range e.Fields {
		if len(field.Name) == 0 {
			panic(errors.New("field name can't be empty"))
		}

		if field.Name == "_id" || field.Name == "id" {
			panic(errors.New("field name _id/id is reserved and can't be used"))
		}

		if field.Name[:1] == "_" {
			panic(errors.New("field name can't start with an underscore"))
		}

		err := e.SetField(field)
		if err != nil {
			return e, err
		}
	}

	// if got write rule, then set add, edit and delete rules for that
	if rule, ok := e.Rules[Write]; ok {
		e.Rules[Add] = rule
		e.Rules[Edit] = rule
		e.Rules[Delete] = rule
	}

	// set default rules
	for _, scope := range scopes {
		if _, ok := e.Rules[scope]; !ok {
			e.Rules[scope] = Admin
		}
	}

	// if private, has to have CreatedBy
	if e.Private {
		if _, ok := e.fields[CreatedBy.Name]; !ok {
			return e, errors.New("private entity has no createdBy field")
		}
	}

	if e.Protected {
		if _, ok := e.fields[PasswordField.Name]; !ok {
			return e, errors.New("password protected entity has no password field")
		}
	}

	return e, nil
}

var CreatedAt = &Field{
	Name:       "createdAt",
	Hidden:     true,
	IsRequired: true,
	Type:       DateTime,
	ValueFunc: func() interface{} {
		return time.Now().UTC()
	},
}

var UpdatedAt = &Field{
	Name:       "updatedAt",
	Hidden:     true,
	IsRequired: true,
	Type:       DateTime,
	ValueFunc: func() interface{} {
		return time.Now().UTC()
	},
}

var PublishedAt = &Field{
	Name:       "publishedAt",
	Hidden:     true,
	IsRequired: true,
	Type:       DateTime,
	TransformFunc: func(ctx *ValueContext, value interface{}) (interface{}, error) {
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
	},
}

var CreatedBy = &Field{
	Name:       "createdBy",
	IsRequired: true,
	Hidden:     true,
	Type:       Key,
	Entity:     User,
	ContextFunc: func(ctx Context) interface{} {
		if len(ctx.User) > 0 {
			if key, err := datastore.DecodeKey(ctx.User); err == nil {
				return key
			}
			return nil
		}
		return nil
	},
}

var UpdatedBy = &Field{
	Name:       "updatedBy",
	IsRequired: true,
	Hidden:     true,
	Type:       Key,
	Entity:     User,
	ContextFunc: func(ctx Context) interface{} {
		if len(ctx.User) > 0 {
			if key, err := datastore.DecodeKey(ctx.User); err == nil {
				return key
			}
			return nil
		}
		return nil
	},
}

func (a *SDK) EnableEntity(e *Entity) (*Entity, error) {
	if len(e.Name) == 0 {
		return e, errors.New("entity name can't be empty")
	}
	if !govalidator.IsAlpha(e.Name) {
		return e, errors.New("entity name can only be a-Z characters")
	}

	e, err := e.init()
	if err != nil {
		return e, err
	}

	//a.enableEntityAPI(e)

	return e, nil
}

func (e *Entity) SetField(field *Field) error {
	if len(field.Name) == 0 {
		return errors.New("field name can't be empty")
	}

	if field.Name == "id" {
		return errors.New("field name 'id' is reserved")
	}

	if e.fields == nil {
		e.fields = map[string]*Field{}
	}

	e.fields[field.Name] = field

	if field.IsRequired {
		e.requiredFields = append(e.requiredFields, field)
	}

	if field.DefaultValue != nil {
		e.preparedData[field] = func(ctx Context, f *Field) interface{} {
			return f.DefaultValue
		}
	}

	if field.ValueFunc != nil {
		e.preparedData[field] = func(ctx Context, f *Field) interface{} {
			return f.ValueFunc()
		}
	}

	if field.ContextFunc != nil {
		e.preparedData[field] = func(ctx Context, f *Field) interface{} {
			return f.ContextFunc(ctx)
		}
	}

	if len(field.ValidateRgx) > 0 {
		field.fieldFunc = append(field.fieldFunc, func(c *ValueContext, v interface{}) (interface{}, error) {
			if c.Trust == High {
				return v, nil
			}

			var matched bool
			var err error

			switch val := v.(type) {
			case string:
				matched, err = regexp.Match(field.ValidateRgx, []byte(val))
				break
			default:
				return v, fmt.Errorf(ErrFieldValueNotValid, c.Field.Name)
			}

			if err != nil {
				return nil, err
			}
			if matched {
				return v, nil
			}

			return v, fmt.Errorf(ErrFieldValueNotValid, c.Field.Name)
		})
	}

	if field.Validator != nil {
		field.fieldFunc = append(field.fieldFunc, func(c *ValueContext, v interface{}) (interface{}, error) {
			if c.Trust == High {
				return v, nil
			}

			ok := c.Field.Validator(v)
			if ok {
				return v, nil
			}
			return v, fmt.Errorf(ErrFieldValueNotValid, c.Field.Name)
		})
	}

	if field.TransformFunc != nil {
		field.fieldFunc = append(field.fieldFunc, field.TransformFunc)
	}

	// if got write rule, then has also add, edit and delete rule
	if rule, ok := field.Rules[Write]; ok {
		field.Rules[Add] = rule
		field.Rules[Edit] = rule
		field.Rules[Delete] = rule
	}

	return nil
}

/**
Adds index document definition and subscribes it to data changes
*/
func (e *Entity) AddIndex(dd *DocumentDefinition) {
	if e.indexes == nil {
		e.indexes = map[string]*DocumentDefinition{}
	}
	e.indexes[dd.Name] = dd
}

var putToIndex = delay.Func(RandStringBytesMaskImprSrc(16), func(ctx context.Context, dd DocumentDefinition, id string, data Data) {
	dd.Put(ctx, id, flatOutput(id, data))
})
var removeFromIndex = delay.Func(RandStringBytesMaskImprSrc(16), func(ctx context.Context, dd DocumentDefinition) {
	// do something expensive!
})

func (e *Entity) PutToIndexes(ctx context.Context, id string, data *DataHolder) {
	for _, dd := range e.indexes {
		err := putToIndex.Call(ctx, *dd, id, data.data)
		if err != nil {
			log.Errorf(ctx, "%v", err.Error())
		}
	}
}
func (e *Entity) RemoveFromIndexes(ctx context.Context) {
	for _, dd := range e.indexes {
		removeFromIndex.Call(ctx, *dd)
	}
}

func (e *Entity) New(ctx Context) (*DataHolder, error) {
	var dataHolder = &DataHolder{
		Entity:  e,
		context: ctx,
		data:    Data{},
		input:   map[string]interface{}{},
		isNew:   true,
	}

	// copy prepared values
	for field, fun := range e.preparedData {
		dataHolder.data[field] = fun(ctx, field)
	}

	if e.OnInit != nil {
		if err := e.OnInit(ctx, dataHolder); err != nil {
			return dataHolder, err
		}
	}

	return dataHolder, nil
}

var (
	ErrKeyNameIdNil         = errors.New("key nameId is nil")
	ErrKeyNameIdInvalidType = errors.New("key nameId invalid type (only string/int64)")
)

func (e *Entity) DecodeKey(c Context, encodedKey string) (Context, *datastore.Key, error) {
	var key *datastore.Key
	var err error

	key, err = datastore.DecodeKey(encodedKey)
	if err != nil {
		return c, key, err
	}

	return c, key, err
}

func (e *Entity) NewIncompleteKey(c Context) (Context, *datastore.Key) {
	var key *datastore.Key

	key = datastore.NewIncompleteKey(c.Context, e.Name, nil)

	return c, key
}

// Gets appengine context and datastore key with optional namespace. It doesn't fail if request is not authenticated.
func (e *Entity) NewKey(c Context, nameId interface{}) (Context, *datastore.Key, error) {
	var key *datastore.Key
	var err error

	if nameId == nil {
		return c, key, ErrKeyNameIdNil
	}

	if stringId, ok := nameId.(string); ok {
		key = datastore.NewKey(c.Context, e.Name, stringId, 0, nil)
	} else if intId, ok := nameId.(int64); ok {
		key = datastore.NewKey(c.Context, e.Name, "", intId, nil)
	} else {
		return c, key, ErrKeyNameIdInvalidType
	}

	return c, key, err
}
