package cms

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"time"
)

func AuthMiddleware(signingKey []byte) *JWTMiddleware {
	return New(MiddlewareOptions{
		Extractor: FromFirst(
			FromAuthHeader,
			FromParameter("token"),
		),
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return signingKey, nil
		},
		SigningMethod:       jwt.SigningMethodHS256,
		CredentialsOptional: true,
	})
}

type Token struct {
	ID      string `json:"id"`
	Expires int64  `json:"expires"`
}

var (
	ErrIllegalAction = errors.New("illegal action")
)

func (c *Context) NewUserToken(userKey string, userRole Role) error {
	var err error
	c.Token, err = newToken(userKey, userRole, c.r.Context().Value("key"))
	return err
}

func newToken(userKey string, userRole Role, privateKey interface{}) (Token, error) {
	var tkn Token

	if len(userKey) == 0 {
		return tkn, ErrIllegalAction
	}

	var exp = time.Now().Add(time.Hour * 12).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"aud": "api",
		"nbf": time.Now().Add(-time.Minute).Unix(),
		"exp": exp,
		"iat": time.Now().Unix(),
		"iss": "sdk",
		"sub": userKey,
		"rol": userRole,
	})

	signed, err := token.SignedString(privateKey)
	if err != nil {
		return tkn, err
	}

	return Token{signed, exp}, nil
}
