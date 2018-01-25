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
)

type Context struct {
	r                *http.Request
	IsAuthenticated  bool
	HasProjectAccess bool
	context.Context
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
	} else {
		ctx.HasProjectAccess = false
		ctx.User = ""
		ctx.Project = ""
		ctx.Context = nil
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

type Result struct {
	Status int         `json:"status"`
	Result interface{} `json:"result"`

	Error   int    `json:"error"`
	Message string `json:"message"`

	Token *Token `json:"token"`
	User  *User  `json:"user"`
}

func (ctx *Context) PrintResult(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")

	var out = Result{
		Status: http.StatusOK,
		Result: result,
	}

	json.NewEncoder(w).Encode(out)
}

func (ctx *Context) PrintAuth(w http.ResponseWriter, user *User, token *Token) {
	w.Header().Set("Content-Type", "application/json")

	var out = Result{
		Status: http.StatusOK,
		User:   user,
		Token:  token,
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

func (ctx *Context) PrintFormError(w http.ResponseWriter, err *Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	var out = Result{
		Status:  http.StatusBadRequest,
		Error:   err.Code,
		Message: err.Message,
	}

	json.NewEncoder(w).Encode(out)
}
