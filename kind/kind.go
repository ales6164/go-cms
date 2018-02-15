package kind

import (
	"errors"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"github.com/asaskevich/govalidator"
)

type Kind struct {
	Name   string  `json:"name"` // Only a-Z characters allowed
	Fields []*Field `json:"fields"`

	subKinds []*Kind // kinds managed by fields
	fields   map[string]*Field
}

type Field struct {
	Name       string
	IsRequired bool
	Multiple   bool
	NoIndex    bool

	isNested bool
	Worker
}

func New(name string, fields []*Field) *Kind {
	if !govalidator.IsAlpha(name) {
		panic(errors.New("kind name must contain a-zA-Z characters only"))
	}
	k := new(Kind)
	k.Name = name
	k.Fields = fields
	for _, f := range fields {
		if len(f.Name) == 0 {
			panic(errors.New("field name can't be empty"))
		}
		if f.Name == "meta" || f.Name == "id" {
			panic(errors.New("field name '" + f.Name + "' already exists"))
		}
		if f.Name[:1] == "_" {
			panic(errors.New("field name can't begin with an underscore"))
		}
		if split := strings.Split(f.Name, "."); len(split) > 1 {
			if split[0] == "meta" || split[0] == "id" {
				panic(errors.New("field name '" + f.Name + "' already exists"))
			}
			f.isNested = true
		}
		if k.fields == nil {
			k.fields = map[string]*Field{}
		}
		k.fields[f.Name] = f
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

func (k *Kind) SubKinds() []*Kind {
	return k.subKinds
}

func (k *Kind) NewIncompleteKey(c context.Context, parent *datastore.Key) *datastore.Key {
	return datastore.NewIncompleteKey(c, k.Name, parent)
}

func (k *Kind) NewKey(c context.Context, nameId string) *datastore.Key {
	return datastore.NewKey(c, k.Name, nameId, 0, nil)
}
