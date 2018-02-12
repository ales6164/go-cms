package api

import (
	"github.com/ales6164/go-cms/entity"
	"github.com/asaskevich/govalidator"
	"net/http"
)

func (a *App) KindSaveDraftHandler(e *entity.Entity) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r)

		h, err := e.NewFromBody(ctx)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		res, err := govalidator.ValidateStruct(&h.Data)
		if err != nil || !res {
			ctx.PrintError(w, err)
			return
		}

		_, err = e.Add(h, entity.StatusDraft)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		ctx.PrintResult(w, h.Data)
	}
}

func (a *App) KindGetHandler(e *entity.Entity) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//ctx := NewContext(r)

		/*vars := mux.Vars(r)
		id := vars["id"]

		key, err := datastore.DecodeKey(id)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}*/

		/*h, err := k.Get(ctx, key)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}*/

		//ctx.PrintResult(w, k)
	}
}

/*

func (a *App) KindUpdateHandler(k *kind.Kind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, h, err := NewContext(r).Parse(k)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		key, err := datastore.DecodeKey(id)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		err = h.Update(key)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		ctx.PrintResult(w, h.Output())
	}
}

func (a *App) KindDeleteHandler(k *kind.Kind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, h, err := NewContext(r).Parse(k)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		key, err := datastore.DecodeKey(id)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		err = h.Delete(key)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		ctx.PrintResult(w, h.Output())
	}
}
*/
