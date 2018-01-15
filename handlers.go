package api

import (
	"net/http"
	"google.golang.org/appengine"
	"io/ioutil"
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"fmt"
	"errors"
)

func OutputData(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, map[string]interface{}{
		"status": http.StatusOK,
		"result": data,
	})
}

func OutputError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, map[string]interface{}{
		"status":  http.StatusInternalServerError,
		"message": err.Error(),
	})
}

func OutputFormError(w http.ResponseWriter, err Error) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, map[string]interface{}{
		"status":  http.StatusBadRequest,
		"error":   err.Code,
		"message": err.Message,
	})
}

func LoginHandler(app *App) func(w http.ResponseWriter, r *http.Request) {
	type Input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Project  string `json:"project"`
	}
	type Output struct {
		*User
		Token string `json:"token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			OutputError(w, err)
			return
		}
		defer r.Body.Close()

		var loginForm Input
		err = json.Unmarshal(body, &loginForm)
		if err != nil {
			OutputError(w, err)
			return
		}

		if !govalidator.IsEmail(loginForm.Email) {
			OutputFormError(w, ErrInvalidEmail)
			return
		}

		if len(loginForm.Password) < 6 {
			OutputFormError(w, ErrPasswordTooShort)
			return
		}

		ctx := appengine.NewContext(r)

		tkn, usr, err := LoginEndpoint(ctx, loginForm.Email, loginForm.Password, loginForm.Project)
		if err != nil {
			OutputError(w, err)
			return
		}

		signedToken, err := tkn.SignedString(app.PrivateKey)
		if err != nil {
			OutputError(w, err)
			return
		}

		OutputData(w, Output{
			User:  usr,
			Token: signedToken,
		})
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
	type Output struct {
		*User
		Token string `json:"token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			OutputError(w, err)
			return
		}
		defer r.Body.Close()

		var input Input
		err = json.Unmarshal(body, &input)
		if err != nil {
			OutputError(w, err)
			return
		}

		if !govalidator.IsEmail(input.Email) {
			OutputFormError(w, ErrInvalidEmail)
			return
		}

		if len(input.Password) < 6 {
			OutputFormError(w, ErrPasswordTooShort)
			return
		}

		ctx := appengine.NewContext(r)

		tkn, usr, err := RegisterEndpoint(ctx, input.Email, input.Password, input.FirstName, input.LastName, input.Avatar)
		if err != nil {
			OutputError(w, err)
			return
		}

		signedToken, err := tkn.SignedString(app.PrivateKey)
		if err != nil {
			OutputError(w, err)
			return
		}

		OutputData(w, Output{
			User:  usr,
			Token: signedToken,
		})
	}
}

var ErrUnathorized = errors.New("unathorized")
func AddProjectHandler(app *App) func(w http.ResponseWriter, r *http.Request) {
	type Input struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}
	type Output struct {
		*User
		Token string `json:"token"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r).WithBody()

		if !ctx.IsAuthenticated {
			ctx.PrintError(w, ErrUnathorized, http.StatusUnauthorized)
			return
		}

		var input Input
		err := json.Unmarshal(ctx.body, &input)
		if err != nil {
			OutputError(w, err)
			return
		}

		tkn, usr, err := RegisterEndpoint(ctx, input.Email, input.Password, input.FirstName, input.LastName, input.Avatar)
		if err != nil {
			OutputError(w, err)
			return
		}

		signedToken, err := tkn.SignedString(app.PrivateKey)
		if err != nil {
			OutputError(w, err)
			return
		}

		OutputData(w, Output{
			User:  usr,
			Token: signedToken,
		})
	}
}
