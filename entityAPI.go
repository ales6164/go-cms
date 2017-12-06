package cms

import (
	"github.com/gorilla/mux"
	"net/http"
	"path"
)

func (e *Entity) handler(p string) http.Handler {
	r := mux.NewRouter()

	joined := path.Join(p, e.Name)

	//r.HandleFunc(name + "/{encodedKey}", e.handleGet()).Methods(http.MethodGet)
	r.HandleFunc(joined+"/{encodedKey}", e.handleUpdate()).Methods(http.MethodPut)
	r.HandleFunc(joined, e.handleAdd()).Methods(http.MethodPost)
	r.HandleFunc(joined, func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r)

		ctx.Print(w, "Hello world")
	}).Methods(http.MethodGet)

	return r
}

func (e *Entity) handleGetEntityInfo() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r).WithBody()
		ctx.Print(w, e)
	}
}

/*func (e *Entity) handleGet() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r).WithScope(Read)
		vars := mux.Vars(r)

		encodedKey := vars["encodedKey"]

		key, err := datastore.DecodeKey(encodedKey)
		if err != nil {
			ctx.PrintError(w, err, http.StatusBadRequest)
			return
		}

		dataHolder, err := e.Get(ctx, key)
		if err != nil {
			ctx.PrintError(w, err, http.StatusInternalServerError)
			return
		}

		ctx.Print(w, dataHolder.Output(ctx))
	}
}*/

func (e *Entity) handleAdd() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r).WithScope(Add)

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

		ctx.Print(w, holder.Output(ctx))
	}
}

func (e *Entity) handleUpdate() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r).WithScope(Edit)
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

		ctx.Print(w, holder.Output(ctx))
	}
}
