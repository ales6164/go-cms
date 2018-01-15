package api

import (
	"time"
	"golang.org/x/crypto/bcrypt"
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

func LoginEndpoint(ctx context.Context, email string, password string, projectEncodedKey string) (*jwt.Token, *User, error) {
	usrKey := datastore.NewKey(ctx, "User", email, 0, nil)

	usr := new(User)
	err := datastore.Get(ctx, usrKey, usr)
	if err != nil {
		return nil, nil, err
	}

	err = decrypt(usr.Hash, []byte(password))
	if err != nil {
		return nil, nil, err
	}

	_, err = getProjectRole(ctx, usrKey, projectEncodedKey)
	if err != nil {
		// return token but without project access
		return newToken(usrKey.Encode(), ""), usr, nil
	} else {
		return newToken(usrKey.Encode(), projectEncodedKey), usr, nil
	}
}

var ErrUserAlreadyExists = errors.New("user with that email already exists")

func RegisterEndpoint(ctx context.Context, email, password, firstName, lastName, avatar string) (*jwt.Token, *User, error) {
	// create password hash
	hash, err := crypt([]byte(password))
	if err != nil {
		return nil, nil, err
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
		err := datastore.Get(ctx, usrKey, nil)

		if err != nil && err == datastore.ErrNoSuchEntity {
			// register
			_, err := datastore.Put(ctx, usrKey, usr)
			return err
		}

		return ErrUserAlreadyExists
	}, nil)
	if err != nil {
		return nil, nil, err
	}

	return newToken(usrKey.Encode(), ""), usr, nil
}

func getProjectRole(ctx context.Context, usrKey *datastore.Key, projectEncodedKey string) (string, error) {
	proKey, err := datastore.DecodeKey(projectEncodedKey)
	if err != nil {
		return "", err
	}

	proAccKey := datastore.NewKey(ctx, "ProjectAccess", proKey.StringID(), 0, usrKey)

	proAcc := new(ProjectAccess)
	err = datastore.Get(ctx, proAccKey, proAcc)
	if err != nil {
		return "", err
	}

	return proAcc.Role, nil
}

func newToken(userKey string, projectKey string) *jwt.Token {
	var exp = time.Now().Add(time.Hour * 12).Unix()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"aud": "api",
		"nbf": time.Now().Add(-time.Minute).Unix(),
		"exp": exp,
		"iat": time.Now().Unix(),
		"iss": "sdk",
		"sub": userKey,
		"pro": projectKey,
	})
	//return token.SignedString(privateKey)
}

func decrypt(hash []byte, password []byte) error {
	defer clear(password)
	return bcrypt.CompareHashAndPassword(hash, password)
}

func crypt(password []byte) ([]byte, error) {
	defer clear(password)
	return bcrypt.GenerateFromPassword(password, 13)
}

func clear(b []byte) {
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
}
