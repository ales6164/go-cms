package cms

import (
	"github.com/dgrijalva/jwt-go"
	gctx "github.com/gorilla/context"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"io/ioutil"
	"net/http"
	"time"
	"google.golang.org/appengine/datastore"
)

type Context struct {
	api *API
	r   *http.Request
	err error

	Context context.Context

	user  User
	token string

	body *Body
}

type User struct {
	userGroup       string
	encodedUserKey  string
	userKey         *datastore.Key
	isAuthenticated bool
}

type Body struct {
	hasReadBody bool
	body        []byte
}

func (ctx Context) User() *datastore.Key {
	return ctx.user.userKey
}
func (ctx Context) UserGroup() string {
	return ctx.user.userGroup
}
func (ctx Context) IsAuthenticated() bool {
	return ctx.user.isAuthenticated
}
func (ctx Context) Token() string {
	return ctx.token
}

func (a *API) NewContext(r *http.Request) Context {
	user, renewedToken := a.getReqUser(r)
	return Context{
		r:       r,
		Context: appengine.NewContext(r),
		user:    user,
		token:   renewedToken,
		body:    &Body{hasReadBody: false},
	}
}

func (ctx Context) WithBody() Context {
	if !ctx.body.hasReadBody {
		ctx.body.body, _ = ioutil.ReadAll(ctx.r.Body)
		ctx.r.Body.Close()
		ctx.body.hasReadBody = true
	}
	return ctx
}

// return true if userKey matches with authenticated user
func (ctx Context) UserMatches(userKey interface{}) bool {
	if ctx.IsAuthenticated() {

		if userKeyString, ok := userKey.(string); ok {
			var err error
			userKey, err = datastore.DecodeKey(userKeyString)
			if err != nil {
				return false
			}
		}

		if userKeyDs, ok := userKey.(*datastore.Key); ok {
			return userKeyDs.Equal(ctx.user.userKey)
		}
	}
	return false
}

func (a *API) getReqUser(r *http.Request) (User, string) {
	var user = User{
		userGroup: "public",
	}
	var userGroup string
	var encodedUserKey string
	var isAuthenticated bool

	var isExpired bool
	var signedRenewedToken string

	tkn := gctx.Get(r, "user")

	if tkn != nil {
		token := tkn.(*jwt.Token)

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

			if err := claims.Valid(); err == nil {
				userGroup = claims["grp"].(string)
				encodedUserKey = claims["sub"].(string)
				isAuthenticated = true
			} else if exp, ok := claims["exp"].(float64); ok {
				// check if it's less than a week old
				if time.Now().Unix()-int64(exp) < time.Now().Add(time.Hour * 24 * 7).Unix() {
					userGroup = claims["grp"].(string)
					encodedUserKey = claims["sub"].(string)
					isAuthenticated = true
					isExpired = true
				}
			}

			// check if everything seems alright
			if isAuthenticated && len(encodedUserKey) > 0 && len(userGroup) > 0 {

				// issue a new token
				if isExpired {
					var err error
					signedRenewedToken, err = _token(encodedUserKey, userGroup, a.signingKey)
					if err != nil {
						return user, signedRenewedToken
					}
				}

				// decode user key for later use
				userKey, err := datastore.DecodeKey(encodedUserKey)
				if err != nil {
					return user, signedRenewedToken
				}

				return User{
					userGroup:       userGroup,
					encodedUserKey:  encodedUserKey,
					userKey:         userKey,
					isAuthenticated: isAuthenticated,
				}, signedRenewedToken
			}
		}
	}

	return user, signedRenewedToken
}
