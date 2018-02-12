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
	Version   int64          `json:"version" datastore:"version"` // 0, 1, 2, 3 ...
	Status    string         `json:"status" datastore:"status"`   // draft, pendingApproval, published, removed
	Label     string         `json:"label" datastore:"label"`     // don't know yet
}

type status string

const (
	StatusDraft     status = "draft"
	StatusPending   status = "pending"
	StatusPublished status = "published"
	StatusRemoved   status = "removed"
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
	return datastore.SaveStruct(h.Data)
}
