package entity

import (
	"time"
	"google.golang.org/appengine/datastore"
	"encoding/json"
	"github.com/ales6164/go-cms/instance"
	"github.com/fatih/structs"
)

type Holder struct {
	entity  *Entity
	context instance.Context

	key        *datastore.Key
	Data       *Data
	properties []datastore.Property
}

type Data struct {
	Meta  Meta        `json:"meta" datastore:"meta"`
	Value interface{} `json:"value" datastore:"value"`
}

type Meta struct {
	CreatedAt time.Time      `json:"createdAt" datastore:"createdAt"`
	CreatedBy *datastore.Key `json:"createdBy" datastore:"createdBy"`
	Label     string         `json:"label" datastore:"label"` // draft, pending, published, removed
}

type label string

const (
	Draft     label = "draft"
	Pending   label = "pending"
	Published label = "published"
	Removed   label = "removed"
)

func (h *Holder) Parse(body []byte) error {
	h.Data = &Data{}
	return json.Unmarshal(body, &h.Data.Value)
}

func (h *Holder) Load(p []datastore.Property) error {
	h.properties = p
	return datastore.LoadStruct(h.Data, p)
}

func (h *Holder) Save() ([]datastore.Property, error) {
	/*var p []datastore.Property

	s := structs.New(&h.Data.Value)
	m := s.Map()
	for key, val := range m {
		field := s.Field(key)
		name := field.Tag("datastore")
		var is
		if len(name) == 0 {
			name = field.Name()
		} else {
			split :=
		}
	}


*/
	return datastore.SaveStruct(&h.Data.Value)
}
