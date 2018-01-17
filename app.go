package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/dgrijalva/jwt-go"
)

type App struct {
	PrivateKey []byte
}

func NewApp() *App {
	a := &App{
		PrivateKey: securecookie.GenerateRandomKey(64),
	}

	return a
}

func (a *App) Serve(rootPath string) {
	r := mux.NewRouter().PathPrefix(rootPath).Subrouter()

	// User handling
	r.HandleFunc("/auth/login", LoginHandler(a)).Methods(http.MethodPost)
	r.HandleFunc("/auth/register", RegisterHandler(a)).Methods(http.MethodPost)

	// Project handling
	r.HandleFunc("/project", NewProjectHandler(a)).Methods(http.MethodPost) // ADD
	r.HandleFunc("/project/{{namespace}}", SelectProjectHandler(a)).Methods(http.MethodPost)  // AUTHORIZE W/ PROJECT

	// Entity handling
	r.HandleFunc("/entity", LoginHandler(a)).Methods(http.MethodPost)   // ADD
	r.HandleFunc("/entity", LoginHandler(a)).Methods(http.MethodPut)    // UPDATE
	r.HandleFunc("/entity", LoginHandler(a)).Methods(http.MethodDelete) // DELETE
	r.HandleFunc("/entity/{{name}}", LoginHandler(a)).Methods(http.MethodGet)    // GET

	// Entity API

	http.Handle(rootPath, &Server{r})
}

func (a *App) SignToken(token *jwt.Token) (*Token, error) {
	signedToken, err := token.SignedString(a.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &Token{Id: signedToken, ExpiresAt: token.Claims.(jwt.MapClaims)["exp"].(int64)}, nil
}
