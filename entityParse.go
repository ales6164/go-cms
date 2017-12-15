package cms

import (
	"encoding/json"
	"reflect"
	"fmt"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"errors"
	"time"
)

func ParseBody(c Context) (map[string]interface{}, error) {
	c = c.WithBody()

	var t map[string]interface{}
	err := json.Unmarshal(c.body.body, &t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func sth() {
	data := `{"name":"Annabella", "age": 22, "hobbies": ["climbing", "playing tennis", 444], "sister": {"name": "Britney"}}`
	var t map[string]interface{}
	json.Unmarshal([]byte(data), &t)

	if _, ok := t["name"].(string); ok {
		fmt.Println("name is string")
	}

	if _, ok := t["age"].(float64); ok {
		fmt.Println("age is float64")
	}

	if _, ok := t["hobbies"].([]interface{}); ok {
		fmt.Println("hobbies is slice")
	}

	if _, ok := t["sister"].(map[string]interface{}); ok {
		fmt.Println("sister is map")
	} else {
		fmt.Println("sister is rather " + reflect.TypeOf(t["sister"]).String())
	}

	fmt.Println(t)
}

var NameFuncMaxRetries = 5

// todo: IDFunc
func (e *Entity) Add(ctx Context, m map[string]interface{}) (*DataHolder, error) {

	dataHolder, err := e.New(ctx).Prepare(m)
	if err != nil {
		return dataHolder, err
	}

	err = datastore.RunInTransaction(ctx.Context, func(tc context.Context) error {

		var key *datastore.Key

		if e.NameFunc != nil {
			if e.nameProvider != nil && dataHolder.nameProviderValue == nil {
				return fmt.Errorf("name value is not provided from '%s' field", e.nameProvider.Name)
			}

			var tempEnt datastore.PropertyList

			for i := 1; i <= NameFuncMaxRetries; i++ {
				var keyName, err = e.NameFunc(dataHolder.nameProviderValue, "", i-1)
				if err != nil {
					return err
				}
				if len(keyName) == 0 {
					return errors.New("name can't be empty")
				}
				key = e.NewKey(ctx, keyName)
				err = datastore.Get(tc, key, &tempEnt)
				if err == nil || err != datastore.ErrNoSuchEntity {
					if i == NameFuncMaxRetries {
						return fmt.Errorf("name function reached max retries with no result")
					}
					continue
				}
				dataHolder.SetProperty(datastore.Property{
					Name:  "meta.name",
					Value: keyName,
				})
				break
			}
		} else {
			key = e.NewIncompleteKey(ctx)
		}

		var now = time.Now()
		dataHolder.SetProperty(datastore.Property{
			Name:  "meta.createdAt",
			Value: now,
		})
		dataHolder.SetProperty(datastore.Property{
			Name:  "meta.updatedAt",
			Value: now,
		})
		if ctx.IsAuthenticated && len(ctx.User) > 0 {
			createdBy, err := datastore.DecodeKey(ctx.User)
			if err != nil {
				return errors.New("error decoding user key")
			}
			dataHolder.SetProperty(datastore.Property{
				Name:  "meta.createdBy",
				Value: createdBy,
			})
		}

		dataHolder.key, err = datastore.Put(tc, key, dataHolder)

		return err
	}, &datastore.TransactionOptions{XG: true})

	return dataHolder, err
}

func (e *Entity) Update(ctx Context, id string, name string, m map[string]interface{}) (*DataHolder, error) {
	var key *datastore.Key
	var err error

	if len(id) > 0 {
		key, err = datastore.DecodeKey(id)
		if err != nil {
			return nil, err
		}
	} else if len(name) > 0 {
		key = e.NewKey(ctx, name)
	} else {
		return nil, errors.New("no id defined")
	}

	dataHolder, err := e.New(ctx).Prepare(m)
	if err != nil {
		return dataHolder, errors.New("prepare: " + err.Error())
	}

	err = datastore.RunInTransaction(ctx.Context, func(tc context.Context) error {
		var newKey *datastore.Key // could also update key if provided with new name

		err := datastore.Get(tc, key, dataHolder)
		if err != nil {
			return err
		}

		if e.NameFunc != nil && e.nameProvider != nil && dataHolder.nameProviderValue != nil {
			var tempEnt datastore.PropertyList

			for i := 1; i <= NameFuncMaxRetries; i++ {
				var keyName, err = e.NameFunc(dataHolder.nameProviderValue, key.StringID(), i)
				if err != nil {
					return err
				}
				if len(keyName) == 0 {
					return errors.New("name can't be empty")
				}

				// skip check if new name matches old
				if key.StringID() != keyName {
					newKey = e.NewKey(ctx, keyName)
					err := datastore.Get(tc, newKey, &tempEnt)
					if err == nil || err != datastore.ErrNoSuchEntity {
						if i == NameFuncMaxRetries {
							return fmt.Errorf("name function reached max retries with not result")
						}
						continue
					}
				}

				dataHolder.SetProperty(datastore.Property{
					Name:  "meta.name",
					Value: keyName,
				})
				break
			}

			// delete old entry
			err = datastore.Delete(tc, key)
			if err != nil {
				return err
			}

			key = newKey
		}

		var now = time.Now()
		dataHolder.SetProperty(datastore.Property{
			Name:  "meta.updatedAt",
			Value: now,
		})
		if ctx.IsAuthenticated && len(ctx.User) > 0 {
			createdBy, err := datastore.DecodeKey(ctx.User)
			if err != nil {
				return errors.New("error decoding user key")
			}
			dataHolder.SetProperty(datastore.Property{
				Name:  "meta.createdBy",
				Value: createdBy,
			})
		}

		dataHolder.key, err = datastore.Put(tc, key, dataHolder)

		return err
	}, &datastore.TransactionOptions{XG: true})

	return dataHolder, err
}
