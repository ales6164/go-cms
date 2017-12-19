package cms

import (
	"errors"
	"google.golang.org/appengine/datastore"
	"net/http"
	"strings"
)

type Entity struct {
	Label   string `json:"label"`   // Only a-Z characters allowed
	Name    string `json:"name"`    // Only a-Z characters allowed
	Private bool   `json:"private"` // Protects entity with user field - only creator has access
	/*Protected bool   `json:"protected"`*/ // Protects entity with password
	Cache bool `json:"cache"`               // Keeps values in memcache - good for categories, translations, ...

	fields map[string]*Field
	Fields []*Field `json:"fields"`

	requiredFields []*Field

	indexes map[string]*DocumentDefinition

	// Listener
	// todo: change this a bit
	OnInit        func(c Context, h *DataHolder) error `json:"-"`
	OnBeforeWrite func(c Context, h *DataHolder) error `json:"-"`
	OnAfterRead   func(c Context, h *DataHolder) error `json:"-"`
	OnAfterWrite  func(c Context, h *DataHolder) error `json:"-"`

	Handler http.Handler
}

func (e *Entity) init() (*Entity, error) {
	for _, field := range e.Fields {

		err := e.SetField(field)
		if err != nil {
			return e, err
		}
	}

	return e, nil
}

func (e *Entity) SetField(field *Field) error {
	if len(field.Name) == 0 {
		panic(errors.New("field name can't be empty"))
	}

	if field.Name == "meta" || field.Name == "id" {
		panic(errors.New("field name '" + field.Name + "' already exists"))
	}

	if field.Name[:1] == "_" {
		panic(errors.New("field name can't begin with an underscore"))
	}

	if split := strings.Split(field.Name, "."); len(split) > 1 {
		if split[0] == "meta" || split[0] == "id" {
			panic(errors.New("field name '" + field.Name + "' already exists"))
		}

		field.isNesting = true
	}

	if e.fields == nil {
		e.fields = map[string]*Field{}
	}

	e.fields[field.Name] = field

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

func (e *Entity) NewIncompleteKey(c Context) *datastore.Key {
	return datastore.NewIncompleteKey(c.Context, e.Name, nil)
}

// Gets appengine context and datastore key with optional namespace. It doesn't fail if request is not authenticated.
func (e *Entity) NewKey(c Context, nameId string) *datastore.Key {
	return datastore.NewKey(c.Context, e.Name, nameId, 0, nil)
}
