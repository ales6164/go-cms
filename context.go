package api

import (
	"golang.org/x/net/context"
	gcontext "github.com/gorilla/context"
	"google.golang.org/appengine"
	"io/ioutil"
	"net/http"
	"github.com/dgrijalva/jwt-go"
	"time"
	"google.golang.org/appengine/datastore"
	"encoding/json"
	"google.golang.org/appengine/log"
)

type Context struct {
	r                *http.Request
	IsAuthenticated  bool
	context.Context
	projectAccessKey *datastore.Key
	userKey          *datastore.Key
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

// Authenticates user; if token is expired, returns a renewed unsigned *jwt.Token
func (ctx Context) Authenticate() (bool, Context, *jwt.Token) {
	var isAuthenticated, isExpired, hasProjectAccessKey bool
	var encUserKey, encProjectAccessKey string
	var unsignedToken *jwt.Token
	var userKey, projectAccessKey *datastore.Key
	var err error

	tkn := gcontext.Get(ctx.r, "auth")
	log.Debugf(ctx, "Authenticate")

	if tkn != nil {
		log.Debugf(ctx, "has token")

		token := tkn.(*jwt.Token)

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

			if err := claims.Valid(); err == nil {
				if encProjectAccessKey, ok = claims["pro"].(string); ok && len(encProjectAccessKey) > 0 {
					hasProjectAccessKey = true
				}
				if encUserKey, ok = claims["sub"].(string); ok && len(encUserKey) > 0 {
					isAuthenticated = true
				}
			} else if exp, ok := claims["exp"].(float64); ok {
				// check if it's less than a week old
				if time.Now().Unix()-int64(exp) < time.Now().Add(time.Hour * 24 * 7).Unix() {
					if encProjectAccessKey, ok = claims["pro"].(string); ok && len(encProjectAccessKey) > 0 {
						hasProjectAccessKey = true
					}
					if encUserKey, ok = claims["sub"].(string); ok && len(encUserKey) > 0 {
						isAuthenticated = true
						isExpired = true
					}
				}
			}
		}
	}

	if hasProjectAccessKey {
		projectAccessKey, err = datastore.DecodeKey(encProjectAccessKey)
		if err != nil {
			// project access key was provided but decoding returned error ... something is not right
			// todo: log this kind of behavior
			isAuthenticated = false
		}
	}

	// check if everything seems alright
	if isAuthenticated {
		// issue a new token
		if isExpired {
			unsignedToken = newToken(encUserKey, encProjectAccessKey)
		}
		userKey, err = datastore.DecodeKey(encUserKey)
		if err != nil {
			// TODO: ERROR COULD INDICATE A HACK - LOG THIS AND NOTIFY ADMIN
			isAuthenticated = false
			unsignedToken = nil
		}
	}

	ctx.IsAuthenticated = isAuthenticated
	ctx.userKey = userKey
	ctx.projectAccessKey = projectAccessKey

	return isAuthenticated, ctx, unsignedToken
}

func newToken(encUserKey string, encProjectAccessKey string) *jwt.Token {
	var exp = time.Now().Add(time.Hour * 72).Unix()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"aud": "api",
		"nbf": time.Now().Add(-time.Minute).Unix(),
		"exp": exp,
		"iat": time.Now().Unix(),
		"iss": "sdk",
		"sub": encUserKey,
		"pro": encProjectAccessKey,
	})
}

/**
RESPONSE
 */

type Result struct {
	Status int         `json:"status"`
	Result interface{} `json:"result"`

	Error   int    `json:"error"`
	Message string `json:"message"`

	Token   *Token   `json:"token"`
	User    *User    `json:"user"`
	Project *Project `json:"project"`
}

func (ctx *Context) PrintResult(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")

	var out = Result{
		Status: http.StatusOK,
		Result: result,
	}

	json.NewEncoder(w).Encode(out)
}

func (ctx *Context) PrintAuth(w http.ResponseWriter, user *User, project *Project, token *Token) {
	w.Header().Set("Content-Type", "application/json")

	var out = Result{
		Status:  http.StatusOK,
		User:    user,
		Token:   token,
		Project: project,
	}

	json.NewEncoder(w).Encode(out)
}

func (ctx *Context) PrintError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	var out = Result{
		Status:  http.StatusInternalServerError,
		Message: err.Error(),
	}

	json.NewEncoder(w).Encode(out)
}

func (ctx *Context) PrintAuthError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	var out = Result{
		Status:  http.StatusUnauthorized,
		Message: "unauthorized",
	}

	json.NewEncoder(w).Encode(out)
}

func (ctx *Context) PrintFormError(w http.ResponseWriter, err Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	var out = Result{
		Status:  http.StatusBadRequest,
		Error:   err.Code,
		Message: err.Message,
	}

	json.NewEncoder(w).Encode(out)
}
