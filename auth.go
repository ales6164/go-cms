package api

import (
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"github.com/dgrijalva/jwt-go"
	"errors"
)

// namespace is email
type User struct {
	Hash      []byte `datastore:"hash,noindex" json:"-"`
	Email     string `datastore:"email" json:"email"`
	FirstName string `datastore:"firstName" json:"firstName"`
	LastName  string `datastore:"lastName" json:"lastName"`
	Avatar    string `datastore:"avatar,noindex" json:"avatar"`
}

type Token struct {
	Id        string `json:"id"`
	ExpiresAt int64  `json:"expiresAt"`
}

func LoginEndpoint(ctx context.Context, email string, password string, projectAccessEncodedKey string) (*jwt.Token, *User, *ProjectAccess, error) {
	usrKey := datastore.NewKey(ctx, "User", email, 0, nil)

	usr := new(User)
	err := datastore.Get(ctx, usrKey, usr)
	if err != nil {
		return nil, nil, nil, err
	}

	err = decrypt(usr.Hash, []byte(password))
	if err != nil {
		return nil, nil, nil, err
	}

	proAccess, err := getProjectAccess(ctx, usrKey, projectAccessEncodedKey)
	if err != nil || len(proAccess.Role) == 0 {
		// return token but without project access
		return newToken(usrKey.Encode(), ""), usr, proAccess, nil
	} else {
		return newToken(usrKey.Encode(), projectAccessEncodedKey), usr, proAccess, nil
	}
}

var ErrUserAlreadyExists = errors.New("user with that email already exists")

func RegisterEndpoint(ctx context.Context, email, password, firstName, lastName, avatar string) (*jwt.Token, *User, *ProjectAccess, error) {
	// create password hash
	hash, err := crypt([]byte(password))
	if err != nil {
		return nil, nil, nil, err
	}

	usr := &User{
		Email:     email,
		Hash:      hash,
		Avatar:    avatar,
		FirstName: firstName,
		LastName:  lastName,
	}

	// create User
	usrKey := datastore.NewKey(ctx, "User", email, 0, nil)
	err = datastore.RunInTransaction(ctx, func(tc context.Context) error {
		err := datastore.Get(ctx, usrKey, &datastore.PropertyList{})

		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				// register
				_, err := datastore.Put(ctx, usrKey, usr)
				return err
			}
		} else {
			return ErrUserAlreadyExists
		}

		return err
	}, nil)
	if err != nil {
		return nil, nil, nil, err
	}

	return newToken(usrKey.Encode(), ""), usr, nil, nil
}

func SelectProjectEndpoint(ctx Context, projectAccessKey *datastore.Key) *jwt.Token {
	return newToken(ctx.userKey.Encode(), projectAccessKey.Encode())
}

func getProjectAccess(ctx context.Context, usrKey *datastore.Key, projectAccessEncodedKey string) (*ProjectAccess, error) {
	proKey, err := datastore.DecodeKey(projectAccessEncodedKey)
	if err != nil {
		return nil, err
	}

	proAccKey := datastore.NewKey(ctx, "ProjectAccess", proKey.StringID(), 0, usrKey)

	proAcc := new(ProjectAccess)
	err = datastore.Get(ctx, proAccKey, proAcc)
	if err != nil {
		return nil, err
	}

	return proAcc, nil
}
