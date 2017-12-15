package cms

import (
	"github.com/gorilla/mux"
	"net/http"
	"google.golang.org/appengine/datastore"
)

func (e *Entity) handleGetEntityInfo() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r).WithBody()
		ctx.Print(w, e)
	}
}

func (e *Entity) handleGet() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, err := NewContext(r).WithEntityAction(e, Read)
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

func (e *Entity) handleAdd() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, err := NewContext(r).WithEntityAction(e, Add)
		if err != nil {
			ctx.PrintError(w, err, http.StatusForbidden)
			return
		}

		data, err := ParseBody(ctx)
		if err != nil {
			ctx.PrintError(w, err, http.StatusBadRequest)
			return
		}

		holder, err := e.Add(ctx, data)
		if err != nil {
			ctx.PrintError(w, err, http.StatusInternalServerError)
			return
		}

		ctx.Print(w, holder.Output(ctx, true))
	}
}

func (e *Entity) handleUpdate() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, err := NewContext(r).WithEntityAction(e, Update)
		if err != nil {
			ctx.PrintError(w, err, http.StatusForbidden)
			return
		}

		vars := mux.Vars(r)
		encodedKey := vars["encodedKey"]

		data, err := ParseBody(ctx)
		if err != nil {
			ctx.PrintError(w, err, http.StatusBadRequest)
			return
		}

		holder, err := e.Update(ctx, encodedKey, "", data)
		if err != nil {
			ctx.PrintError(w, err, http.StatusInternalServerError)
			return
		}

		ctx.Print(w, holder.Output(ctx, true))
	}
}
