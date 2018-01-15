package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
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
	r.HandleFunc("/project", LoginHandler(a)).Methods(http.MethodPost) // ADD
	r.HandleFunc("/project", LoginHandler(a)).Methods(http.MethodGet) // AUTHORIZE W/ PROJECT

	// Entity handling
	r.HandleFunc("/entity", LoginHandler(a)).Methods(http.MethodPost) // ADD
	r.HandleFunc("/entity", LoginHandler(a)).Methods(http.MethodPut) // UPDATE
	r.HandleFunc("/entity", LoginHandler(a)).Methods(http.MethodDelete) // DELETE
	r.HandleFunc("/entity", LoginHandler(a)).Methods(http.MethodGet) // GET

	// Entity API

	http.Handle(rootPath, r)
}