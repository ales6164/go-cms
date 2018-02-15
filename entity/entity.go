package entity

import (
	"reflect"
	"google.golang.org/appengine/datastore"
	"github.com/ales6164/go-cms/instance"
)

/*type PropertyLoadSaver interface {
	Load([]Property) error
	Save() ([]Property, error)
}*/

type Entity struct {
	Name string
	Type reflect.Type
}

func (e *Entity) New(ctx instance.Context) *Holder {
	h := &Holder{
		entity:  e,
		context: ctx,
		Data: &Data{
			Value: reflect.New(e.Type).Interface(),
		},
	}
	if id := ctx.Id(); len(id) != 0 {
		h.key, _ = datastore.DecodeKey(id)
	}
	return h
}

func (e *Entity) NewFromBody(ctx instance.Context) (*Holder, error) {
	h := e.New(ctx)
	err := h.Parse(ctx.Body())
	return h, err
}
