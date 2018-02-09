package entity

import (
	"github.com/ales6164/go-cms"
	"reflect"
	"google.golang.org/appengine/datastore"
)

/*type PropertyLoadSaver interface {
	Load([]Property) error
	Save() ([]Property, error)
}*/

type Entity struct {
	Name string
	Type reflect.Type
}

func (e *Entity) New(ctx api.Context) *Holder {
	id := ctx.Id()
	var key *datastore.Key
	if len(id) > 0 {
		key, _ = datastore.DecodeKey(id)
	}
	return &Holder{
		entity:  e,
		hasKey:  key != nil,
		key:     key,
		context: ctx,
		Value:   reflect.New(e.Type).Interface(),
	}
}

func (e *Entity) NewFromBody(ctx api.Context) (*Holder, error) {
	h := e.New(ctx)
	err := h.Parse(ctx.Body())
	return h, err
}

//TODO:
func (e *Entity) SaveDraft(h *Holder) (*datastore.Key, error) {
	h.status = "draft"

	if h.hasKey {
		//TODO: change status of the old draft to draftOld
		//TODO: put everything inside a transaction
	} else {
		h.key = datastore.NewIncompleteKey(h.context, e.Name, nil)
	}

	return datastore.Put(h.context, h.key, h)
}
