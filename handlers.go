package api

import (
	"net/http"
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"google.golang.org/appengine/datastore"
	"strings"
	"golang.org/x/net/context"
)

func (a *App) AuthLoginHandler() http.HandlerFunc {
	type Input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r).WithBody()

		var input Input
		err := json.Unmarshal(ctx.body, &input)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		input.Email = strings.ToLower(input.Email)

		// verify input
		if !govalidator.IsEmail(input.Email) || len(input.Email) < 6 || len(input.Email) > 64 {
			ctx.PrintFormError(w, ErrInvalidEmail)
			return
		}
		if len(input.Password) < 6 || len(input.Password) > 128 {
			ctx.PrintFormError(w, ErrPasswordLength)
			return
		}

		// get user
		userKey := datastore.NewKey(ctx, "User", input.Email, 0, nil)
		user := new(User)
		err = datastore.Get(ctx, userKey, user)
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				ctx.PrintFormError(w, ErrUserDoesNotExist)
				return
			}
			ctx.PrintError(w, err)
			return
		}

		// decrypt hash
		err = decrypt(user.Hash, []byte(input.Password))
		if err != nil {
			ctx.PrintFormError(w, ErrUserPasswordIncorrect)
			// todo: log and report
			return
		}

		// get user projects
		user.Projects, _ = GetUserProjects(ctx, userKey)

		// create a token
		token := newToken(user.Email, "")

		// sign the new token
		signedToken, err := a.SignToken(token)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		ctx.PrintAuth(w, user, signedToken)
	}
}

func (a *App) AuthRegistrationHandler() http.HandlerFunc {
	type Input struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Photo     string `json:"photo"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r).WithBody()

		var input Input
		err := json.Unmarshal(ctx.body, &input)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		input.Email = strings.ToLower(input.Email)

		// verify input
		if !govalidator.IsEmail(input.Email) || len(input.Email) < 6 || len(input.Email) > 64 {
			ctx.PrintFormError(w, ErrInvalidEmail)
			return
		}
		if len(input.Password) < 6 || len(input.Password) > 128 {
			ctx.PrintFormError(w, ErrPasswordLength)
			return
		}
		if len(input.Photo) > 0 && !govalidator.IsURL(input.Photo) {
			ctx.PrintFormError(w, ErrPhotoInvalidFormat)
			return
		}

		// create password hash
		hash, err := crypt([]byte(input.Password))
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		// create User
		user := &User{
			Email:     input.Email,
			Hash:      hash,
			Photo:     input.Photo,
			FirstName: input.FirstName,
			LastName:  input.LastName,
		}

		err = datastore.RunInTransaction(ctx, func(tc context.Context) error {
			userKey := datastore.NewKey(tc, "User", user.Email, 0, nil)
			err := datastore.Get(tc, userKey, &datastore.PropertyList{})
			if err != nil {
				if err == datastore.ErrNoSuchEntity {
					// register
					_, err := datastore.Put(tc, userKey, user)
					return err
				}
				return err
			}
			return ErrUserAlreadyExists
		}, nil)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		// create a token
		token := newToken(user.Email, "")

		// sign the new token
		signedToken, err := a.SignToken(token)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		ctx.PrintAuth(w, user, signedToken)
	}
}

// Renews token and fetches user info
func (a *App) AuthRenewProjectAccessTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, renewedToken := NewContext(r).renew()

		if renewedToken == nil {
			ctx.PrintAuthError(w)
			return
		}

		// get user
		userKey := datastore.NewKey(ctx, "User", ctx.User, 0, nil)
		user := new(User)
		err := datastore.Get(ctx, userKey, user)
		if err != nil {
			ctx.PrintAuthError(w)
			return
		}

		// check project access
		if ctx.HasProjectAccess {
			proAccessKey := datastore.NewKey(ctx, "ProjectAccess", ctx.Project, 0, userKey)
			proAccess := new(ProjectAccess)
			err := datastore.Get(ctx, proAccessKey, proAccess)
			if err != nil {
				ctx.PrintAuthError(w)
				return
			}
		}

		// get user projects
		user.Projects, _ = GetUserProjects(ctx, userKey)

		// sign the new token
		signedToken, err := a.SignToken(renewedToken)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		ctx.PrintAuth(w, user, signedToken)
	}
}

func (a *App) CreateProjectHandler() http.HandlerFunc {
	type Input struct {
		Name      string `valid:"length(4|32),required" json:"name"`
		Namespace string `valid:"length(4|32),isSlug,required" json:"namespace"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		authenticated, ctx := NewContext(r).WithBody().Authenticate(false)

		if !authenticated {
			ctx.PrintAuthError(w)
			return
		}

		input := new(Input)
		err := json.Unmarshal(ctx.body, input)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		// verify input
		ok, err := govalidator.ValidateStruct(input)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}
		if !ok {
			ctx.PrintFormError(w, ErrInvalidFormInput)
			return
		}

		userKey := datastore.NewKey(ctx, "User", ctx.User, 0, nil)
		user := new(User)
		err = datastore.Get(ctx, userKey, user)
		if err != nil {
			ctx.PrintAuthError(w)
			return
		}

		proKey := datastore.NewKey(ctx, "Project", input.Namespace, 0, nil)
		proAccessKey := datastore.NewKey(ctx, "ProjectAccess", input.Namespace, 0, userKey)

		pro := &Project{
			Name:      input.Name,
			Namespace: input.Namespace,
		}

		proAccess := &ProjectAccess{
			Project: proKey,
			User:    userKey,
			Role:    "admin",
		}

		err = datastore.RunInTransaction(ctx, func(tc context.Context) error {
			err := datastore.Get(tc, proKey, &datastore.PropertyList{})
			if err != nil {
				if err == datastore.ErrNoSuchEntity {
					_, err = datastore.Put(tc, proKey, pro)
					if err == nil {
						_, err = datastore.Put(tc, proAccessKey, proAccess)
					}
				}
				return err
			}
			return ErrProjectAlreadyExists
		}, &datastore.TransactionOptions{
			XG: true,
		})

		token := newToken(user.Email, pro.Namespace)

		// get user projects
		user.Projects, _ = GetUserProjects(ctx, userKey)

		// sign the new token
		signedToken, err := a.SignToken(token)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		ctx.PrintAuth(w, user, signedToken)
	}
}
