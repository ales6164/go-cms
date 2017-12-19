package cms

import (
	"google.golang.org/appengine/datastore"
	"time"
)

// PreparedEntity data holder
type DataHolder struct {
	context Context
	Entity  *Entity `json:"entity"`

	key *datastore.Key

	request             *APIRequest
	preparedInputData   map[*Field][]datastore.Property // user input
	hasLoadedStoredData bool
	loadedStoredData    map[string][]datastore.Property // data already stored in datastore - if exists
	savedData           []datastore.Property            // list of properties stored to datastore

	isOldVersion bool // when updating entity we want to also update old entry meta.

	nameProviderValue interface{} // value for entity.NameFunc
}

func (e *Entity) New(ctx Context) *DataHolder {
	var dataHolder = &DataHolder{
		Entity:            e,
		context:           ctx,
		preparedInputData: map[*Field][]datastore.Property{},
		loadedStoredData:  map[string][]datastore.Property{},
	}

	return dataHolder
}

func (holder *DataHolder) Output(ctx Context, lookup bool) map[string]interface{} {
	var output = map[string]interface{}{}

	// range over data. Value can be single value or if the field it Multiple then it's an array
	for _, prop := range holder.savedData {
		output = holder.appendPropertyValue(output, prop, holder.Entity.fields[prop.Name], lookup)
	}

	output["id"] = holder.key.Encode()

	return output
}

func (holder *DataHolder) Load(ps []datastore.Property) error {
	holder.hasLoadedStoredData = true
	for _, prop := range ps {
		holder.loadedStoredData[prop.Name] = appendProperty(holder.loadedStoredData[prop.Name], prop)
	}
	return nil
}

func (holder *DataHolder) Save() ([]datastore.Property, error) {
	var ps []datastore.Property

	if holder.isOldVersion {
		holder.loadedStoredData["meta.status"] = []datastore.Property{{
			Name:  "meta.status",
			Value: nil,
		}}

		for _, psArr := range holder.loadedStoredData {
			ps = append(holder.savedData, psArr...)
		}
	} else {
		// check if required fields are there
		for _, field := range holder.Entity.Fields {
			var inputProperties = holder.preparedInputData[field]
			var loadedProperties = holder.loadedStoredData[field.Name]

			if len(inputProperties) != 0 {
				holder.savedData = append(holder.savedData, inputProperties...)
			} else if len(loadedProperties) != 0 {
				holder.savedData = append(holder.savedData, loadedProperties...)
			} else if field.IsRequired {
				return nil, FormError{FormErrFieldRequired, field}
			}

		}

		// set meta tags
		var now = time.Now()
		holder.savedData = append(holder.savedData, datastore.Property{
			Name:  "meta.updatedAt",
			Value: now,
		})
		holder.savedData = append(holder.savedData, datastore.Property{
			Name:  "meta.status",
			Value: "active",
		})
		if holder.hasLoadedStoredData {
			if metaCreatedAt, ok := holder.loadedStoredData["meta.createdAt"]; ok {
				holder.savedData = append(holder.savedData, metaCreatedAt[0])
			}
			if metaCreatedBy, ok := holder.loadedStoredData["meta.createdBy"]; ok {
				holder.savedData = append(holder.savedData, metaCreatedBy[0])
			}
			if metaVersion, ok := holder.loadedStoredData["meta.version"]; ok {
				metaVersion[0].Value = metaVersion[0].Value.(int64) + 1
				holder.savedData = append(holder.savedData, metaVersion[0])
			}
		} else {
			holder.savedData = append(holder.savedData, datastore.Property{
				Name:  "meta.version",
				Value: int64(0),
			})
		}
		if holder.context.IsAuthenticated() {
			holder.savedData = append(holder.savedData, datastore.Property{
				Name:  "meta.updatedBy",
				Value: holder.context.User(),
			})
		}

		ps = holder.savedData
	}

	return ps, nil
}
