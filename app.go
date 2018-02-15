package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/dgrijalva/jwt-go"
	"github.com/ales6164/go-cms/middleware"
	"github.com/asaskevich/govalidator"
	"github.com/ales6164/go-cms/kind"
	"strings"
	"github.com/ales6164/go-cms/instance"
)

type App struct {
	PrivateKey []byte
	Kinds      []*kind.Kind
	kinds      map[string]*kind.Kind
}

func NewApp() *App {
	a := &App{
		//PrivateKey: securecookie.GenerateRandomKey(64),
		PrivateKey: []byte("MVoBOkxWGi7pwM1bN9hgxgEVjVXmhTAq"),
		kinds:      map[string]*kind.Kind{},
	}

	govalidator.CustomTypeTagMap.Set("isSlug", govalidator.CustomTypeValidator(IsSlug))

	return a
}

/*func (a *App) Import(class interface{}) {
	classType := getType(class)
	name := classType.Name()
	if len(name) < 3 {
		panic(errors.New("entity name is too short"))
	}
	if name == "auth" || name == "project" {
		panic(errors.New("entity name can't contain '" + name + "'"))
	}
	if !govalidator.IsAlpha(name) {
		panic(errors.New("entity name can only contain a-zA-Z characters"))
	}

	a.entities = append(a.entities, &entity.Entity{Name: name, Type: classType})
	*//*a.Kinds = append(a.Kinds, class)
	a.kinds[class.Name] = class*//*
}*/

func (a *App) Import(kind *kind.Kind) {
	a.Kinds = append(a.Kinds, kind)
	a.kinds[kind.Name] = kind
}

/*
Only have custom API defined kinds
 */
func (a *App) Serve(rootPath string) {
	authMiddleware := middleware.AuthMiddleware(a.PrivateKey)
	r := mux.NewRouter().PathPrefix(rootPath).Subrouter()

	// Create project kind
	//r.Handle("/api/{project}", authMiddleware.Handler(KindHandler(a))).Methods(http.MethodPost)

	//r.HandleFunc("/api", a.GetKindDefinitions()).Methods(http.MethodGet)

	// Create project
	r.Handle("/project", authMiddleware.Handler(a.CreateProjectHandler())).Methods(http.MethodPost)

	// User authorization
	r.HandleFunc("/auth/login", a.AuthLoginHandler()).Methods(http.MethodPost)
	r.HandleFunc("/auth/register", a.AuthRegistrationHandler()).Methods(http.MethodPost)

	// Project/User re-authorization
	r.Handle("/auth", authMiddleware.Handler(a.AuthRenewProjectAccessTokenHandler())).Methods(http.MethodPost)
	r.Handle("/auth/{project}", authMiddleware.Handler(a.AuthRenewProjectAccessTokenHandler())).Methods(http.MethodPost)

	// API
	for _, ent := range a.kinds {
		name := strings.ToLower(ent.Name)
		r.Handle("/"+name, authMiddleware.Handler(a.GetHandler(ent))).Methods(http.MethodGet)

		//r.Handle("/"+name+"/draft", authMiddleware.Handler(a.AddDraftHandler(ent))).Methods(http.MethodPost) // ADD
		r.Handle("/"+name, authMiddleware.Handler(a.AddHandler(ent))).Methods(http.MethodPost)               // ADD
		//r.Handle("/"+name+"/{id}", authMiddleware.Handler(a.KindGetHandler(e))).Methods(http.MethodGet)       // GET
		//r.Handle("/{project}/api/"+name+"/{id}", authMiddleware.Handler(a.KindUpdateHandler(e))).Methods(http.MethodPut)    // UPDATE
		//r.Handle("/{project}/api/"+name+"/{id}", authMiddleware.Handler(a.KindDeleteHandler(e))).Methods(http.MethodDelete) // DELETE
	}

	http.Handle(rootPath, &Server{r})
}

func (a *App) SignToken(token *jwt.Token) (*instance.Token, error) {
	signedToken, err := token.SignedString(a.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &instance.Token{Id: signedToken, ExpiresAt: token.Claims.(jwt.MapClaims)["exp"].(int64)}, nil
}
