package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/appengine/datastore"
)

func (ctx Context) newToken(userKey *datastore.Key, userGroup string, privateKey interface{}) (Context, error) {
	var encodedUserKey = userKey.Encode()
	var err error

	ctx.token, err = _token(encodedUserKey, userGroup, privateKey)
	if err != nil {
		return ctx, err
	}

	ctx.user = User{
		userGroup:       userGroup,
		userKey:         userKey,
		encodedUserKey:  encodedUserKey,
		isAuthenticated: true,
	}

	return ctx, nil
}

func _token(encodedUserKey string, userGroup string, privateKey interface{}) (string, error) {
	var exp = time.Now().Add(time.Hour * 12).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"aud": "api",
		"nbf": time.Now().Add(-time.Minute).Unix(),
		"exp": exp,
		"iat": time.Now().Unix(),
		"iss": "sdk",
		"sub": encodedUserKey,
		"grp": userGroup,
	})
	return token.SignedString(privateKey)
}
