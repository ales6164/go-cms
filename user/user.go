package user

import (
	"context"

	"google.golang.org/appengine/datastore"
)

type UserAccount struct {
	Email     string `datastore:"email"`
	AccessKey string `datastore:"accessKey" json:"-"`
	UserGroup string `datastore:"userGroup"`
	Hash      []byte `datastore:"hash,noindex" json:"-"`
}

type Profile struct {
	FirstName string `datastore:"firstName" json:"firstName"`
	LastName  string `datastore:"lastName" json:"lastName"`
	Avatar    string `datastore:"avatar" json:"avatar"`
}

func (a *API) login(ctx Context, username string, password string) (Context, error) {
	key, ua, err := getUser(ctx.Context, username)
	if err != nil {
		return ctx, err
	}

	err = decrypt(ua.Hash, []byte(password))
	if err != nil {
		return ctx, err
	}

	ctx, err = ctx.newToken(key, a.options.DefaultUserGroup, tokenKey)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func getUser(ctx context.Context, username string) (*datastore.Key, *UserAccount, error) {
	var ua = new(UserAccount)

	var key = datastore.NewKey(ctx, "_UserAccount", username, 0, nil)

	err := datastore.Get(ctx, key, ua)
	if err != nil {
		return key, ua, err
	}

	return key, ua, nil
}
