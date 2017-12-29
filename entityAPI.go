package cms

import (
	"github.com/gorilla/mux"
	"net/http"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

/*func (e *Entity) handleGetEntityInfo() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r).WithBody()
		ctx.Print(w, e)
	}
}*/

var (
	FormErrFieldRequired ErrorType = "required"
)

type FormError struct {
	errorType ErrorType
	field     *Field
}

type APIRequest struct {
	Entity        string
	Entry         *datastore.Key `datastore:",noindex"`
	Status        string
	StatusMessage string         `datastore:",noindex"`
	Scope         string

	Method        string `datastore:",noindex"`
	RemoteAddr    string
	Body          []byte `datastore:",noindex"`
	Host          string `datastore:",noindex"`
	Path          string `datastore:",noindex"`
	UserAgent     string `datastore:",noindex"`
	Referer       string `datastore:",noindex"`
	ContentLength int64  `datastore:",noindex"`
	TLSVersion    int16  `datastore:",noindex"`
	CipherSuite   int16  `datastore:",noindex"`
	Proto         string `datastore:",noindex"`
}

type ErrorType string

func (e FormError) Error() string {
	return "form error"
}

func (a *API) handleGet(e *Entity) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, err := a.NewContext(r).HasPermission(e, Read)
		if err != nil {
			ctx.PrintError(w, err, http.StatusForbidden)
			return
		}

		vars := mux.Vars(r)
		encodedKey := vars["encodedKey"]

		key, err := datastore.DecodeKey(encodedKey)
		if err != nil {
			key = e.NewKey(ctx, encodedKey)
		}

		var dataHolder = e.New(ctx)
		dataHolder.key = key

		err = datastore.Get(ctx.Context, key, dataHolder)
		if err != nil {
			ctx.PrintError(w, err, http.StatusInternalServerError)
			return
		}

		ctx.Print(w, dataHolder.Output(ctx, true))
	}
}

func (a *API) handleCreate(e *Entity) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, err := a.NewContext(r).HasPermission(e, Create)
		if err != nil {
			ctx.PrintError(w, err, http.StatusForbidden)
			return
		}

		holder, err := e.New(ctx).PrepareFromInput()
		if err != nil {
			ctx.PrintError(w, err, http.StatusBadRequest)
			return
		}

		holder, err = e.Create(ctx, holder)
		if err != nil {
			ctx.PrintError(w, err, http.StatusInternalServerError)
			return
		}

		ctx.Print(w, holder.Output(ctx, true))
	}
}

func (e *Entity) Create(ctx Context, dataHolder *DataHolder) (*DataHolder, error) {
	var err error

	dataHolder.key = e.NewIncompleteKey(ctx, nil)
	dataHolder.key, err = datastore.Put(ctx.Context, dataHolder.key, dataHolder)
	if err != nil {
		dataHolder.request.Status = "error"
		dataHolder.request.StatusMessage = err.Error()
	} else {
		dataHolder.request.Status = "ok"
	}

	var requestKey = datastore.NewIncompleteKey(ctx.Context, "_APIRequest", dataHolder.key)
	dataHolder.request.Scope = string(Create)
	dataHolder.request.Entity = e.Name
	dataHolder.request.Entry = dataHolder.key

	_, reqErr := datastore.Put(ctx.Context, requestKey, dataHolder.request)
	if reqErr != nil {
		log.Criticalf(ctx.Context, "failure saving _APIRequest error: %s", reqErr.Error())
	}

	dataHolder.updateSearchIndex()

	return dataHolder, err
}

func (a *API) handleUpdate(e *Entity) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, err := a.NewContext(r).HasPermission(e, Update)
		if err != nil {
			ctx.PrintError(w, err, http.StatusForbidden)
			return
		}

		vars := mux.Vars(r)
		encodedKey := vars["encodedKey"]

		holder, err := e.New(ctx).PrepareFromInput()
		if err != nil {
			ctx.PrintError(w, err, http.StatusBadRequest)
			return
		}

		holder.key, err = datastore.DecodeKey(encodedKey)
		if err != nil {
			ctx.PrintError(w, err, http.StatusBadRequest)
			return
		}

		holder, err = e.Update(ctx, holder)
		if err != nil {
			ctx.PrintError(w, err, http.StatusInternalServerError)
			return
		}

		ctx.Print(w, holder.Output(ctx, true))
	}
}

func (e *Entity) Update(ctx Context, dataHolder *DataHolder) (*DataHolder, error) {
	var err error
	err = datastore.RunInTransaction(ctx.Context, func(tc context.Context) error {

		err := datastore.Get(tc, dataHolder.key, dataHolder)
		if err != nil {
			return err
		}

		var replacementKey = dataHolder.Entity.NewIncompleteKey(ctx, dataHolder.key)
		var oldHolder = dataHolder.OldHolder(replacementKey)

		var keys = []*datastore.Key{replacementKey, dataHolder.key}
		var holders = []interface{}{oldHolder, dataHolder}

		keys, err = datastore.PutMulti(ctx.Context, keys, holders)
		if err != nil {
			dataHolder.request.Status = "error"
			dataHolder.request.StatusMessage = err.Error()
		} else {
			dataHolder.request.Status = "ok"
		}

		var requestKey = datastore.NewIncompleteKey(ctx.Context, "_APIRequest", dataHolder.key)
		dataHolder.request.Scope = string(Update)
		dataHolder.request.Entity = e.Name
		dataHolder.request.Entry = dataHolder.key

		_, reqErr := datastore.Put(ctx.Context, requestKey, dataHolder.request)
		if reqErr != nil {
			log.Criticalf(ctx.Context, "failure saving _APIRequest error: %s", reqErr.Error())
		}

		return err
	}, &datastore.TransactionOptions{XG: true})

	if err != nil {
		return dataHolder, err
	}

	dataHolder.updateSearchIndex()

	return dataHolder, err
}
