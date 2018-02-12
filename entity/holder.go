package entity

import (
	"time"
	"google.golang.org/appengine/datastore"
	"github.com/ales6164/go-cms"
	"encoding/json"
)

type Holder struct {
	entity  *Entity
	context api.Context
	key     *datastore.Key
	hasKey  bool
	Value   interface{}
	loaded  []datastore.Property
	status  string
}

func (h *Holder) Parse(body []byte) error {
	return json.Unmarshal(body, &h.Value)
}

func (h *Holder) Load(p []datastore.Property) error {
	h.loaded = append(h.loaded, p...)
	return nil
}

// Save saves all of l's properties as a slice or Properties.
func (h *Holder) Save() ([]datastore.Property, error) {
	var props []datastore.Property

	valueProps, err := datastore.SaveStruct(h.Value)
	if err != nil {
		return nil, err
	}

	props = append(props, datastore.Property{
		Name:  "meta.status",
		Value: h.status,
	})

	props = append(props, datastore.Property{
		Name:  "meta.createdAt",
		Value: time.Now(),
	})

	props = append(props, datastore.Property{
		Name:  "meta.createdBy",
		Value: h.context.UserKey,
	})

	props = append(props, valueProps...)

	return props, nil
}
