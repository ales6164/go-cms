package api

import (
	"time"
	"google.golang.org/appengine/datastore"
	"net/http"
	"github.com/asaskevich/govalidator"
	"encoding/json"
)

type Kind struct {
	Meta Meta `valid:"-" datastore:"meta" json:"meta"`
}

type Meta struct {
	CreatedAt time.Time      `valid:"-" datastore:"createdAt" json:"createdAt"`
	UpdatedAt time.Time      `valid:"-" datastore:"updatedAt" json:"updatedAt"`
	CreatedBy *datastore.Key `valid:"-" datastore:"createdBy" json:"createdBy"`
	UpdatedBy *datastore.Key `valid:"-" datastore:"updatedBy" json:"updatedBy"`
}

type Options struct {
}

func (a *App) KindGetHandler(kind interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authenticated, ctx := NewContext(r).WithBody().Authenticate(true)

		if !authenticated {
			ctx.PrintAuthError(w)
			return
		}

		err := json.Unmarshal(ctx.body, kind)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		_, err = govalidator.ValidateStruct(kind)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}
	}
}
