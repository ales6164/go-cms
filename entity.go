package cms

import (
	"errors"
	"google.golang.org/appengine/datastore"
	"net/http"
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
	NameFunc     func(providedFieldValue interface{}, oldName string, failedCount int) string `json:"-"`
	nameProvider *Field

	preparedData map[*Field]func(ctx Context, f *Field) interface{}

	requiredFields []*Field

	indexes map[string]*DocumentDefinition

	// Rules
	Rules Rules `json:"rules"`

	// Listener
	// todo: change this a bit
	OnInit        func(c Context, h *DataHolder) error `json:"-"`
	OnBeforeWrite func(c Context, h *DataHolder) error `json:"-"`
	OnAfterRead   func(c Context, h *DataHolder) error `json:"-"`
	OnAfterWrite  func(c Context, h *DataHolder) error `json:"-"`

	Handler http.Handler
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

		if field.Name == "id" {
			panic(errors.New("field name 'id' is reserved and can't be used"))
		}

		if field.Name == "meta" {
			panic(errors.New("field name 'meta' is reserved and can't be used"))
		}

		if field.Name[:1] == "_" {
			panic(errors.New("field name can't start with an underscore"))
		}

		err := e.SetField(field)
		if err != nil {
			return e, err
		}

		if field.IsNameProvider {
			if e.nameProvider != nil {
				return e, errors.New("multiple name provider fields detected")
			}
			e.nameProvider = field
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
	if e.Protected {
		if _, ok := e.fields[PasswordField.Name]; !ok {
			return e, errors.New("password protected entity has no password field")
		}
	}

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

	/*if field.ContextFunc != nil {
		e.preparedData[field] = func(ctx Context, f *Field) interface{} {
			return f.ContextFunc(ctx)
		}
	}*/

	/*if len(field.Validate) > 0 {
		field.fieldFunc = append(field.fieldFunc, func( v interface{}) (interface{}, error) {

			var matched bool
			var err error

			switch val := v.(type) {
			case string:
				matched, err = regexp.Match(field.Validate, []byte(val))
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
	}*/

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
/*func (e *Entity) AddIndex(dd *DocumentDefinition) {
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
}*/

var (
	ErrKeyNameIdNil         = errors.New("key nameId is nil")
	ErrKeyNameIdInvalidType = errors.New("key nameId invalid type (only string/int64)")
)

func (e *Entity) NewIncompleteKey(c Context)  *datastore.Key {
	return datastore.NewIncompleteKey(c.Context, e.Name, nil)
}

// Gets appengine context and datastore key with optional namespace. It doesn't fail if request is not authenticated.
func (e *Entity) NewKey(c Context, nameId string)  *datastore.Key {
	return datastore.NewKey(c.Context, e.Name, nameId, 0, nil)
}
