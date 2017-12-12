package cms

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/asaskevich/govalidator"
	"fmt"
	"path"
)

type API struct {
	path          string
	router        *mux.Router
	Handler       http.Handler
	EditorHandler http.Handler
	middleware    *JWTMiddleware
	sessionStore  *sessions.CookieStore
	entities      []*Entity
	installed     bool
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

	r.Handle("/", editor())

	a := &API{
		router:        r,
		middleware:    AuthMiddleware(signingKey),
		Handler:       &Server{r},
		EditorHandler: editor(),
	}

	return a
}

func (a *API) Handle(p string, e *Entity) {

	a.router.HandleFunc(path.Join(p, "{encodedKey}"), e.handleGet()).Methods(http.MethodGet)
	a.router.HandleFunc(path.Join(p, "{encodedKey}"), e.handleUpdate()).Methods(http.MethodPut)
	a.router.HandleFunc(p, e.handleAdd()).Methods(http.MethodPost)
	a.router.HandleFunc(p, func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r)

		ctx.Print(w, "Hello world")
	}).Methods(http.MethodGet)

	return
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
