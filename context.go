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
	r   *http.Request
	err error

	Context context.Context

	User string // encoded User key
	Role Role
	Rank int

	IsAuthenticated bool
	Token           Token

	body *Body
}

type Body struct {
	hasReadBody bool
	body        []byte
}

func NewContext(r *http.Request) Context {
	isAuthenticated, userRole, userKey, renewedToken, err := getUser(r)
	return Context{
		r:               r,
		Context:         appengine.NewContext(r),
		IsAuthenticated: isAuthenticated,
		Role:            userRole,
		Rank:            Ranks[userRole],
		User:            userKey,
		Token:           renewedToken,
		err:             err,
		body:            &Body{hasReadBody: false},
	}
}

func (c Context) WithBody() Context {
	if !c.body.hasReadBody {
		c.body.body, _ = ioutil.ReadAll(c.r.Body)
		c.r.Body.Close()
		c.body.hasReadBody = true
	}
	return c
}

// return true if userKey matches with userKey in token
func (c Context) UserMatches(userKey interface{}) bool {
	if userKeyString, ok := userKey.(string); ok {
		return userKeyString == c.User
	} else if userKeyDs, ok := userKey.(*datastore.Key); ok {
		if key, err := datastore.DecodeKey(c.User); err == nil {
			return userKeyDs.StringID() == key.StringID()
		}
	}
	return false
}

func getUser(r *http.Request) (bool, Role, string, Token, error) {
	var isAuthenticated bool
	var userRole Role = "guest"
	var userKey string
	var renewedToken Token
	var err error

	tkn := gctx.Get(r, "user")

	if tkn != nil {
		token := tkn.(*jwt.Token)

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			err = claims.Valid()
			if err == nil {
				var username string
				if username, ok = claims["sub"].(string); ok {
					if userRole, ok := claims["rol"].(int); ok {
						return true, Role(userRole), username, renewedToken, err
					}
				}
				return isAuthenticated, userRole, userKey, renewedToken, ErrIllegalAction
			} else if exp, ok := claims["exp"].(float64); ok {
				// check if it's less than a week old
				if time.Now().Unix()-int64(exp) < time.Now().Add(time.Hour * 24 * 7).Unix() {
					if userKey, ok := claims["sub"].(string); ok {
						if userRole, ok := claims["rol"].(int); ok {
							renewedToken, err = newToken(userKey, Role(userRole))
							if err != nil {
								return isAuthenticated, Role(userRole), userKey, renewedToken, err
							}
							return true, Role(userRole), userKey, renewedToken, err
						}
					}
				}
			}
		}
	}

	return isAuthenticated, userRole, userKey, renewedToken, err
}
