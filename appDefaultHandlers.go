package api

import (
	"net/http"
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"google.golang.org/appengine/datastore"
	"strings"
	"golang.org/x/net/context"
	"github.com/ales6164/go-cms/user"
	"github.com/ales6164/go-cms/project"
	"github.com/ales6164/go-cms/instance"
)

func (a *App) AuthLoginHandler() http.HandlerFunc {
	type Input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := instance.NewContext(r)

		var input Input
		err := json.Unmarshal(ctx.Body(), &input)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		input.Email = strings.ToLower(input.Email)

		// verify input
		if !govalidator.IsEmail(input.Email) || len(input.Email) < 6 || len(input.Email) > 64 {
			ctx.PrintError(w, instance.ErrInvalidEmail)
			return
		}
		if len(input.Password) < 6 || len(input.Password) > 128 {
			ctx.PrintError(w, instance.ErrPasswordLength)
			return
		}

		// get user
		userKey := datastore.NewKey(ctx, "User", input.Email, 0, nil)
		user := new(user.User)
		err = datastore.Get(ctx, userKey, user)
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				ctx.PrintError(w, instance.ErrUserDoesNotExist)
				return
			}
			ctx.PrintError(w, err)
			return
		}

		// decrypt hash
		err = decrypt(user.Hash, []byte(input.Password))
		if err != nil {
			ctx.PrintError(w, instance.ErrUserPasswordIncorrect)
			// todo: log and report
			return
		}

		// get user projects
		user.Projects, _ = project.GetUserProjects(ctx, userKey)

		// create a token
		token := instance.NewToken(user.Email, "")

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
		ctx := instance.NewContext(r)

		var input Input
		err := json.Unmarshal(ctx.Body(), &input)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		input.Email = strings.ToLower(input.Email)

		// verify input
		if !govalidator.IsEmail(input.Email) || len(input.Email) < 6 || len(input.Email) > 64 {
			ctx.PrintError(w, instance.ErrInvalidEmail)
			return
		}
		if len(input.Password) < 6 || len(input.Password) > 128 {
			ctx.PrintError(w, instance.ErrPasswordLength)
			return
		}
		if len(input.Photo) > 0 && !govalidator.IsURL(input.Photo) {
			ctx.PrintError(w, instance.ErrPhotoInvalidFormat)
			return
		}

		// create password hash
		hash, err := crypt([]byte(input.Password))
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		// create User
		user := &user.User{
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
			return instance.ErrUserAlreadyExists
		}, nil)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		// create a token
		token := instance.NewToken(user.Email, "")

		// sign the new token
		signedToken, err := a.SignToken(token)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		ctx.PrintAuth(w, user, signedToken)
	}
}
