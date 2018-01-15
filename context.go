package api

import (
	"golang.org/x/net/context"
	gcontext "github.com/gorilla/context"
	"google.golang.org/appengine"
	"io/ioutil"
	"net/http"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type Context struct {
	IsAuthenticated bool
	Request         *http.Request
	Context         context.Context
	Token           string
	UnsignedToken   *jwt.Token
	Project         string
	User            string
	*Body
}

type ContextUser struct {
	isExpired       bool
	isAuthenticated bool
	Project         string
	User            string
}

type Body struct {
	hasReadBody bool
	body        []byte
}

func NewContext(r *http.Request) Context {
	user, renewedToken := authorize(r)
	return Context{
		IsAuthenticated: user.isAuthenticated,
		Request:         r,
		Context:         appengine.NewContext(r),
		Project:         user.Project,
		User:            user.User,
		UnsignedToken:   renewedToken,
		Body:            &Body{hasReadBody: false},
	}
}

func (ctx Context) WithBody() Context {
	if !ctx.Body.hasReadBody {
		ctx.Body.body, _ = ioutil.ReadAll(ctx.Request.Body)
		ctx.Request.Body.Close()
		ctx.Body.hasReadBody = true
	}
	return ctx
}

func authorize(r *http.Request) (*ContextUser, *jwt.Token) {
	user := new(ContextUser)

	tkn := gcontext.Get(r, "auth")

	if tkn != nil {
		token := tkn.(*jwt.Token)

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

			if err := claims.Valid(); err == nil {
				user.Project = claims["pro"].(string)
				user.User = claims["sub"].(string)
				user.isAuthenticated = true
			} else if exp, ok := claims["exp"].(float64); ok {
				// check if it's less than a week old
				if time.Now().Unix()-int64(exp) < time.Now().Add(time.Hour * 24 * 7).Unix() {
					user.Project = claims["pro"].(string)
					user.User = claims["sub"].(string)
					user.isAuthenticated = true
					user.isExpired = true
				}
			}

			// check if everything seems alright
			if user.isAuthenticated && len(user.User) > 0 {
				// issue a new token
				if user.isExpired {
					signedRenewedToken := newToken(user.User, user.Project)
					return user, signedRenewedToken

				}
				return user, nil
			}
		}
	}
	return nil, nil
}
