package api

import (
	"net/http"
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
)

func LoginHandler(app *App) func(w http.ResponseWriter, r *http.Request) {
	type Input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Project  string `json:"project"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r).WithBody()

		var input Input
		err := json.Unmarshal(ctx.body, &input)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		if !govalidator.IsEmail(input.Email) {
			ctx.PrintFormError(w, ErrInvalidEmail)
			return
		}

		if len(input.Password) < 6 {
			ctx.PrintFormError(w, ErrPasswordTooShort)
			return
		}

		tkn, usr, proAccess, err := LoginEndpoint(ctx, input.Email, input.Password, input.Project)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		var pro *Project
		if proAccess != nil {
			pro, err = GetProject(ctx, proAccess.Project)
			if err != nil {
				ctx.PrintError(w, err)
				return
			}
		}

		signedToken, err := app.SignToken(tkn)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		ctx.PrintAuth(w, usr, pro, signedToken)
	}
}

func RegisterHandler(app *App) func(w http.ResponseWriter, r *http.Request) {
	type Input struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Avatar    string `json:"avatar"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r).WithBody()

		var input Input
		err := json.Unmarshal(ctx.body, &input)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		if !govalidator.IsEmail(input.Email) {
			ctx.PrintFormError(w, ErrInvalidEmail)
			return
		}

		if len(input.Password) < 6 {
			ctx.PrintFormError(w, ErrPasswordTooShort)
			return
		}

		tkn, usr, _, err := RegisterEndpoint(ctx, input.Email, input.Password, input.FirstName, input.LastName, input.Avatar)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		signedToken, err := app.SignToken(tkn)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		ctx.PrintAuth(w, usr, nil, signedToken)
	}
}

func NewProjectHandler(app *App) func(w http.ResponseWriter, r *http.Request) {
	type Input struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		authenticated, ctx, _ := NewContext(r).WithBody().Authenticate()

		if !authenticated {
			ctx.PrintAuthError(w)
			return
		}

		var input Input
		err := json.Unmarshal(ctx.body, &input)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		_, proAccessKey, pro, _, err := NewProject(ctx, input.Name, input.Namespace)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		tkn := SelectProjectEndpoint(ctx, proAccessKey)

		signedToken, err := app.SignToken(tkn)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		ctx.PrintAuth(w, nil, pro, signedToken)
	}
}

func SelectProjectHandler(app *App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		authenticated, ctx, _ := NewContext(r).WithBody().Authenticate()

		if !authenticated {
			ctx.PrintAuthError(w)
			return
		}

		namespace := mux.Vars(r)["namespace"]
		if len(namespace) == 0 {
			ctx.PrintAuthError(w)
			return
		}

		proAccessKey, proAccess, err := GetProjectAccess(ctx, namespace)
		if err != nil || len(proAccess.Role) == 0 {
			ctx.PrintAuthError(w)
			return
		}

		tkn := SelectProjectEndpoint(ctx, proAccessKey)

		pro, err := GetProject(ctx, proAccess.Project)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		signedToken, err := app.SignToken(tkn)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		ctx.PrintAuth(w, nil, pro, signedToken)
	}
}
