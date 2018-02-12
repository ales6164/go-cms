package entity

import (
	"google.golang.org/appengine/datastore"
	"time"
)

/**
FETCH
 */

// undefined
func (e *Entity) Get(h *Holder) {

}

// undefined
func (e *Entity) Query(h *Holder) {

}

/**
ADD; EDIT; DELETE; PROMOTE
 */

func (e *Entity) Add(h *Holder, s status) (*datastore.Key, error) {
	h.key = datastore.NewIncompleteKey(h.context, e.Name, nil)

	h.Data.Meta = Meta{
		CreatedAt: time.Now(),
		CreatedBy: h.context.UserKey,
		Version:   0,
		Status:    string(s),
	}

	return datastore.Put(h.context, h.key, h.Data)
}

func (e *Entity) Update(h *Holder, key *datastore.Key) {
	h.key = key
}

func (e *Entity) Delete(h *Holder, key *datastore.Key) {
	h.key = key

}

func (e *Entity) Promote(h *Holder, key *datastore.Key, newStatus status) {
	h.key = key

}
