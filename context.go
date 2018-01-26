package api

import (
	"golang.org/x/net/context"
	gcontext "github.com/gorilla/context"
	"google.golang.org/appengine"
	"io/ioutil"
	"net/http"
	"github.com/dgrijalva/jwt-go"
	"time"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/ales6164/go-cms/user"
	"google.golang.org/appengine/datastore"
	"github.com/ales6164/go-cms/kind"
)

type Context struct {
	r                *http.Request
	IsAuthenticated  bool
	HasProjectAccess bool
	context.Context
	userKey          *datastore.Key
	User             string
	Project          string
	*Body
}

type Body struct {
	hasReadBody bool
	body        []byte
}

func NewContext(r *http.Request) Context {
	return Context{
		r:       r,
		Context: appengine.NewContext(r),
		Body:    &Body{hasReadBody: false},
	}
}

func (ctx Context) WithBody() Context {
	if !ctx.Body.hasReadBody {
		ctx.Body.body, _ = ioutil.ReadAll(ctx.r.Body)
		ctx.r.Body.Close()
		ctx.Body.hasReadBody = true
	}
	return ctx
}

func (ctx Context) Parse(k *kind.Kind) (Context, *kind.Holder, error) {
	_, c := ctx.WithBody().Authenticate(true)
	h := k.NewHolder(c, c.userKey)
	return c, h, h.ParseInput(c.body)
}

// Authenticates user; if token is expired, returns a renewed unsigned *jwt.Token
func (ctx Context) Authenticate(requireProjectAccess bool) (bool, Context) {
	var isAuthenticated, isExpired, hasProjectNamespace bool
	var userEmail, projectNamespace string

	tkn := gcontext.Get(ctx.r, "auth")
	if tkn != nil {
		token := tkn.(*jwt.Token)
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if err := claims.Valid(); err == nil {
				if projectNamespace, ok = claims["pro"].(string); ok && len(projectNamespace) > 0 {
					hasProjectNamespace = true
				}
				if userEmail, ok = claims["sub"].(string); ok && len(userEmail) > 0 {
					isAuthenticated = true
				}
			} else if exp, ok := claims["exp"].(float64); ok {
				// check if it's less than a week old
				if time.Now().Unix()-int64(exp) < time.Now().Add(time.Hour * 24 * 7).Unix() {
					if projectNamespace, ok = claims["pro"].(string); ok && len(projectNamespace) > 0 {
						hasProjectNamespace = true
					}
					if userEmail, ok = claims["sub"].(string); ok && len(userEmail) > 0 {
						isAuthenticated = true
						isExpired = true
					}
				}
			}
		}
	}

	ctx.IsAuthenticated = isAuthenticated && (hasProjectNamespace || !requireProjectAccess) && !isExpired
	if ctx.IsAuthenticated {
		ctx.HasProjectAccess = hasProjectNamespace
		ctx.User = userEmail
		ctx.Project = projectNamespace
		ctx.userKey = datastore.NewKey(ctx, "User", userEmail, 0, nil)
	} else {
		ctx.HasProjectAccess = false
		ctx.User = ""
		ctx.Project = ""
	}

	return ctx.IsAuthenticated, ctx
}

// Authenticates user; if token is expired, returns a renewed unsigned *jwt.Token
func (ctx Context) renew() (Context, *jwt.Token) {
	var isAuthenticated, hasProjectNamespace bool
	var userEmail, projectNamespace string
	var unsignedToken *jwt.Token

	tkn := gcontext.Get(ctx.r, "auth")
	if tkn != nil {
		token := tkn.(*jwt.Token)

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

			if err := claims.Valid(); err == nil {
				if projectNamespace, ok = claims["pro"].(string); ok && len(projectNamespace) > 0 {
					hasProjectNamespace = true
				}
				if userEmail, ok = claims["sub"].(string); ok && len(userEmail) > 0 {
					isAuthenticated = true
				}
			} else if exp, ok := claims["exp"].(float64); ok {
				// check if it's less than a week old
				if time.Now().Unix()-int64(exp) < time.Now().Add(time.Hour * 24 * 7).Unix() {
					if projectNamespace, ok = claims["pro"].(string); ok && len(projectNamespace) > 0 {
						hasProjectNamespace = true
					}
					if userEmail, ok = claims["sub"].(string); ok && len(userEmail) > 0 {
						isAuthenticated = true
					}
				}
			}
		}
	}

	ctx.IsAuthenticated = isAuthenticated
	ctx.User = userEmail

	vars := mux.Vars(ctx.r)
	newProjectNamespace := vars["project"]
	if len(newProjectNamespace) > 0 {
		ctx.HasProjectAccess = true
		ctx.Project = newProjectNamespace
	} else {
		ctx.HasProjectAccess = hasProjectNamespace
		ctx.Project = projectNamespace
	}

	// issue a new token
	if isAuthenticated {
		unsignedToken = newToken(ctx.User, ctx.Project)
	}

	return ctx, unsignedToken
}

func newToken(userEmail string, projectNamespace string) *jwt.Token {
	var exp = time.Now().Add(time.Hour * 72).Unix()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"aud": "api",
		"nbf": time.Now().Add(-time.Minute).Unix(),
		"exp": exp,
		"iat": time.Now().Unix(),
		"iss": "sdk",
		"sub": userEmail,
		"pro": projectNamespace,
	})
}

/**
RESPONSE
 */

type Token struct {
	Id        string `json:"id"`
	ExpiresAt int64  `json:"expiresAt"`
}

type AuthResult struct {
	Token *Token     `json:"token"`
	User  *user.User `json:"user"`
}

func (ctx *Context) PrintResult(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	
	json.NewEncoder(w).Encode(result)
}

func (ctx *Context) PrintAuth(w http.ResponseWriter, user *user.User, token *Token) {
	w.Header().Set("Content-Type", "application/json")

	var out = AuthResult{
		User:   user,
		Token:  token,
	}

	json.NewEncoder(w).Encode(out)
}

func (ctx *Context) PrintError(w http.ResponseWriter, err error) {
	if err == ErrUnathorized {
		w.WriteHeader(http.StatusUnauthorized)
	} else if _, ok := err.(*Error); ok {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write([]byte(err.Error()))
}
