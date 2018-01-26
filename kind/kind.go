package kind

import (
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"strings"
	"errors"
)

type Kind struct {
	Name   string   `json:"name"` // Only a-Z characters allowed
	Fields []*Field `json:"fields"`

	fields         map[string]*Field
	requiredFields []*Field
}

func New(name string, fields ...*Field) *Kind {
	k := new(Kind)
	k.Name = name
	for _, field := range fields {
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
		if k.fields == nil {
			k.fields = map[string]*Field{}
		}
		k.fields[field.Name] = field
	}
	return k
}

func (k *Kind) NewHolder(ctx context.Context, user *datastore.Key) *Holder {
	return &Holder{
		Kind:              k,
		context:           ctx,
		user:              user,
		preparedInputData: map[*Field][]datastore.Property{},
		loadedStoredData:  map[string][]datastore.Property{},
	}
}

func (k *Kind) NewIncompleteKey(c context.Context, parent *datastore.Key) *datastore.Key {
	return datastore.NewIncompleteKey(c, k.Name, parent)
}

func (k *Kind) NewKey(c context.Context, nameId string) *datastore.Key {
	return datastore.NewKey(c, k.Name, nameId, 0, nil)
}
