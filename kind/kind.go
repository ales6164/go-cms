package kind

import (
	"errors"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type Kind struct {
	Name   string  `json:"name"` // Only a-Z characters allowed
	Fields []Field `json:"field"`

	fields map[string]Field
}

type Field interface {
	GetName() string
	GetRequired() bool
	GetMultiple() bool
	GetNoIndex() bool
	GetNested() bool

	Init() error
	Parse(value interface{}) ([]datastore.Property, error)
	Output(ctx context.Context, value interface{}) interface{}
}

func New(name string, fields ...Field) *Kind {
	k := new(Kind)
	k.Name = name
	for _, f := range fields {
		if len(f.GetName()) == 0 {
			panic(errors.New("field name can't be empty"))
		}
		if f.GetName() == "meta" || f.GetName() == "id" {
			panic(errors.New("field name '" + f.GetName() + "' already exists"))
		}
		if f.GetName()[:1] == "_" {
			panic(errors.New("field name can't begin with an underscore"))
		}
		if split := strings.Split(f.GetName(), "."); len(split) > 1 {
			if split[0] == "meta" || split[0] == "id" {
				panic(errors.New("field name '" + f.GetName() + "' already exists"))
			}
			if f.GetNested() == false {
				panic(errors.New("field name '" + f.GetName() + "' contains dots but is not nested"))
			}
		}
		if err := f.Init(); err != nil {
			panic(err)
		}
		if k.fields == nil {
			k.fields = map[string]Field{}
		}
		k.fields[f.GetName()] = f
	}
	return k
}

func (k *Kind) NewHolder(ctx context.Context, user *datastore.Key) *Holder {
	return &Holder{
		Kind:              k,
		context:           ctx,
		user:              user,
		preparedInputData: map[Field][]datastore.Property{},
		loadedStoredData:  map[string][]datastore.Property{},
	}
}

func (k *Kind) NewIncompleteKey(c context.Context, parent *datastore.Key) *datastore.Key {
	return datastore.NewIncompleteKey(c, k.Name, parent)
}

func (k *Kind) NewKey(c context.Context, nameId string) *datastore.Key {
	return datastore.NewKey(c, k.Name, nameId, 0, nil)
}
