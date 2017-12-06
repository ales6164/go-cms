package cms

import (
	"time"
)

var CacheExpiration = time.Duration(0)

/*
func (e *Entity) Lookup(ctx Context, id string) (map[string]interface{}, error) {
	var data map[string]interface{}

	_, err := memcache.Gob.Get(ctx.Context, id, &data)
	if err != nil {
		ctx, key, err := e.DecodeKey(ctx, id)
		if err != nil {
			return data, err
		}
		holder, err := e.Get(ctx, key)
		if err != nil {
			return data, err
		}

		data, err = e.cacheData(ctx, holder) // cache for future use
	}

	return data, err
}

func (e *Entity) CacheData(ctx Context, holder *DataHolder) error {
	_, err := e.cacheData(ctx, holder)
	return err
}

func (e *Entity) cacheData(ctx Context, holder *DataHolder) (map[string]interface{}, error) {
	var output = output(ctx, holder.id, holder.data, false)

	var item = &memcache.Item{
		Key:        holder.id,
		Expiration: CacheExpiration,
		Object:     output,
	}

	return output, memcache.Gob.Set(ctx.Context, item)
}
*/
