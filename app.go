package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/dgrijalva/jwt-go"
	"github.com/ales6164/go-cms/middleware"
	"github.com/asaskevich/govalidator"
	"github.com/ales6164/go-cms/kind"
	"google.golang.org/appengine/datastore"
)

type App struct {
	PrivateKey []byte
	Kinds      []*kind.Kind
}

func NewApp() *App {
	a := &App{
		//PrivateKey: securecookie.GenerateRandomKey(64),
		PrivateKey: []byte("MVoBOkxWGi7pwM1bN9hgxgEVjVXmhTAq"),
	}

	govalidator.CustomTypeTagMap.Set("isSlug", govalidator.CustomTypeValidator(IsSlug))

	return a
}

func (a *App) DefineKind(class *kind.Kind) {
	a.Kinds = append(a.Kinds, class)
}

func (a *App) Serve(rootPath string) {
	authMiddleware := middleware.AuthMiddleware(a.PrivateKey)
	r := mux.NewRouter().PathPrefix(rootPath).Subrouter()

	// API
	for _, k := range a.Kinds {
		r.Handle("/api/{project}/"+k.Name, authMiddleware.Handler(a.KindAddHandler(k))).Methods(http.MethodPost)        // ADD
		r.Handle("/api/{project}/"+k.Name+"/{id}", authMiddleware.Handler(a.KindGetHandler(k))).Methods(http.MethodGet) // GET
		r.Handle("/api/{project}/"+k.Name+"/{id}", authMiddleware.Handler(a.KindUpdateHandler(k))).Methods(http.MethodPut) // UPDATE
	}
	// CUSTOM KINDS:
	//r.Handle("/api/{project}/{kind}/{id}", authMiddleware.Handler(APIGetHandler(a))).Methods(http.MethodGet)       // GET
	//r.Handle("/api/{project}/{kind}", authMiddleware.Handler(APIAddHandler(a))).Methods(http.MethodPost)           // ADD
	//r.Handle("/api/{project}/{kind}/{id}", authMiddleware.Handler(APIUpdateHandler(a))).Methods(http.MethodPut)    // UPDATE
	//r.Handle("/api/{project}/{kind}/{id}", authMiddleware.Handler(APIDeleteHandler(a))).Methods(http.MethodDelete) // DELETE

	// Create project kind
	//r.Handle("/api/{project}", authMiddleware.Handler(KindHandler(a))).Methods(http.MethodPost)

	// Create project
	r.Handle("/api", authMiddleware.Handler(a.CreateProjectHandler())).Methods(http.MethodPost)

	// User authorization
	r.HandleFunc("/auth/login", a.AuthLoginHandler()).Methods(http.MethodPost)
	r.HandleFunc("/auth/register", a.AuthRegistrationHandler()).Methods(http.MethodPost)

	// Project/User re-authorization
	r.Handle("/auth", authMiddleware.Handler(a.AuthRenewProjectAccessTokenHandler())).Methods(http.MethodPost)
	r.Handle("/auth/{project}", authMiddleware.Handler(a.AuthRenewProjectAccessTokenHandler())).Methods(http.MethodPost)

	http.Handle(rootPath, &Server{r})
}

func (a *App) SignToken(token *jwt.Token) (*Token, error) {
	signedToken, err := token.SignedString(a.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &Token{Id: signedToken, ExpiresAt: token.Claims.(jwt.MapClaims)["exp"].(int64)}, nil
}

func (a *App) KindAddHandler(k *kind.Kind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, h, err := NewContext(r).Parse(k)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		err = h.Add()
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		ctx.PrintResult(w, h.Output(false))
	}
}

func (a *App) KindGetHandler(k *kind.Kind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r)

		vars := mux.Vars(r)
		id := vars["id"]

		key, err := datastore.DecodeKey(id)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		h, err := k.Get(ctx, key)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		ctx.PrintResult(w, h.Output(false))
	}
}

func (a *App) KindUpdateHandler(k *kind.Kind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, h, err := NewContext(r).Parse(k)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		key, err := datastore.DecodeKey(id)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		err = h.Update(key)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		ctx.PrintResult(w, h.Output(false))
	}
}
