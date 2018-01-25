package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/dgrijalva/jwt-go"
	"github.com/ales6164/go-cms/middleware"
	"github.com/asaskevich/govalidator"
	"strings"
)

type App struct {
	PrivateKey []byte
	Kinds      []interface{}
}

func NewApp() *App {
	a := &App{
		//PrivateKey: securecookie.GenerateRandomKey(64),
		PrivateKey: []byte("MVoBOkxWGi7pwM1bN9hgxgEVjVXmhTAq"),
	}

	govalidator.CustomTypeTagMap.Set("isSlug", govalidator.CustomTypeValidator(IsSlug))

	a.DefineKind((*Kind)(nil))

	return a
}

func (a *App) DefineKind(class interface{}) {
	a.Kinds = append(a.Kinds, class)
}

func (a *App) Serve(rootPath string) {
	authMiddleware := middleware.AuthMiddleware(a.PrivateKey)
	r := mux.NewRouter().PathPrefix(rootPath).Subrouter()

	// API
	for _, kind := range a.Kinds {
		kindName := strings.ToLower(getType(kind))
		r.Handle("/api/{project}/"+kindName+"/{id}", authMiddleware.Handler(a.KindGetHandler(kind))).Methods(http.MethodGet)
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
