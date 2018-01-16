package api

import (
	"golang.org/x/net/context"
)

type Entity struct {
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Multiple bool   `json:"multiple"`
	NoIndex  bool   `json:"noIndex"`
}

func AddEntity(ctx context.Context, e Entity) {

	/*datastore.RunInTransaction(ctx, func(tc context.Context) error {
		key := datastore.NewKey(tc, "Entity", e.Name, )
		datastore.Get(tc, )
	}, nil)*/
}

func UpdateEntity(ctx context.Context, e Entity) {

}

func RemoveEntity(ctx context.Context, e Entity) {

}