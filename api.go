package cms

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/asaskevich/govalidator"
	"fmt"
)

type API struct {
	path          string
	router        *mux.Router
	Handler       http.Handler
	EditorHandler http.Handler
	middleware    *JWTMiddleware
	sessionStore  *sessions.CookieStore

	entities       []*Entity
	handledEntities []*Entity
}

type Server struct {
	h *mux.Router
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Cache-Control, "+
				"X-Requested-With")
	}
	if req.Method == "OPTIONS" {
		return
	}

	s.h.ServeHTTP(w, req)
}

var signingKey []byte

func NewAPI() *API {
	signingKey = securecookie.GenerateRandomKey(64)
	var r = mux.NewRouter()

	a := &API{
		router:     r,
		middleware: AuthMiddleware(signingKey),
		Handler:    &Server{r},
	}

	a.EditorHandler = a.editor()

	return a
}

func (a *API) Handle(p string, e *Entity) *mux.Router {

	var sub = a.router.PathPrefix(p).Subrouter()

	sub.HandleFunc("/{encodedKey}", e.handleGet()).Methods(http.MethodGet)
	sub.HandleFunc("/{encodedKey}", e.handleUpdate()).Methods(http.MethodPut)
	sub.HandleFunc("", e.handleAdd()).Methods(http.MethodPost)
	sub.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r)
		ctx.Print(w, "Hello "+e.Name)
	}).Methods(http.MethodGet)

	a.handledEntities = append(a.handledEntities, e)

	return sub
}

func (a *API) HandleFunc(p string, f func(w http.ResponseWriter, r *http.Request)) *mux.Router {

	var sub = a.router.PathPrefix(p).Subrouter()

	sub.HandleFunc("", f).Methods(http.MethodPost)

	return sub
}

func (a *API) Add(es ...*Entity) error {
	for i, e := range es {
		if len(e.Name) == 0 {
			return fmt.Errorf("adding item %v: entity name can't be empty", i)
		}
		if !govalidator.IsAlpha(e.Name) {
			return fmt.Errorf("adding item %v: entity name must contain only aA-zZ characters", i)
		}

		e, err := e.init()
		if err != nil {
			return err
		}

		a.entities = append(a.entities, e)
	}
	return nil
}
