package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/dgrijalva/jwt-go"
	"github.com/ales6164/go-cms/middleware"
	"github.com/asaskevich/govalidator"
	"github.com/ales6164/go-cms/kind"
	"errors"
	"strings"
	"github.com/ales6164/go-cms/entity"
	"google.golang.org/appengine/datastore"
)

type App struct {
	PrivateKey []byte
	Kinds      []*kind.Kind
	kinds      map[string]*kind.Kind

	entities []*entity.Entity
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

func (a *App) Import(class interface{}) {
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
	/*a.Kinds = append(a.Kinds, class)
	a.kinds[class.Name] = class*/
}

func (a *App) DefineKind(class interface{}) {
	/*a.Kinds = append(a.Kinds, class)
	a.kinds[class.Name] = class*/
}

/*func (a *App) DefineKind(class *kind.Kind) {
	if _, ok := a.kinds[class.Name]; ok {
		panic(fmt.Errorf("kind with name %s already exists", class.Name))
	}
	a.Kinds = append(a.Kinds, class)
	a.kinds[class.Name] = class

	subKinds := class.SubKinds()
	if len(subKinds) > 0 {
		for _, k := range subKinds {
			k.Name = class.Name + "_" + k.Name

			if _, ok := a.kinds[k.Name]; ok {
				panic(fmt.Errorf("kind with name %s already exists", k.Name))
			}
			a.Kinds = append(a.Kinds, k)
			a.kinds[k.Name] = k
		}
	}
}*/

/*
Only have custom API defined kinds
 */
func (a *App) Serve(rootPath string) {
	authMiddleware := middleware.AuthMiddleware(a.PrivateKey)
	r := mux.NewRouter().PathPrefix(rootPath).Subrouter()

	// CUSTOM KINDS:
	/*	r.Handle("/{project}/api/{kind}/{id}", authMiddleware.Handler(APIGetHandler(a))).Methods(http.MethodGet)       // GET
		r.Handle("/{project}/api/{kind}", authMiddleware.Handler(APIAddHandler(a))).Methods(http.MethodPost)           // ADD
		r.Handle("/{project}/api/{kind}/{id}", authMiddleware.Handler(APIUpdateHandler(a))).Methods(http.MethodPut)    // UPDATE
		r.Handle("/{project}/api/{kind}/{id}", authMiddleware.Handler(APIDeleteHandler(a))).Methods(http.MethodDelete) // DELETE*/

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
	for _, ent := range a.entities {
		name := strings.ToLower(ent.Name)
		r.Handle("/"+name, authMiddleware.Handler(a.KindGetHandler(ent))).Methods(http.MethodGet)

		r.Handle("/"+name, authMiddleware.Handler(a.KindSaveDraftHandler(ent))).Methods(http.MethodPost) // ADD
		//r.Handle("/"+name+"/{id}", authMiddleware.Handler(a.KindGetHandler(e))).Methods(http.MethodGet)       // GET
		//r.Handle("/{project}/api/"+name+"/{id}", authMiddleware.Handler(a.KindUpdateHandler(e))).Methods(http.MethodPut)    // UPDATE
		//r.Handle("/{project}/api/"+name+"/{id}", authMiddleware.Handler(a.KindDeleteHandler(e))).Methods(http.MethodDelete) // DELETE
	}

	http.Handle(rootPath, &Server{r})
}

func (a *App) SignToken(token *jwt.Token) (*Token, error) {
	signedToken, err := token.SignedString(a.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &Token{Id: signedToken, ExpiresAt: token.Claims.(jwt.MapClaims)["exp"].(int64)}, nil
}

func (a *App) KindSaveDraftHandler(e *entity.Entity) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r)

		h, err := e.NewFromBody(ctx)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		res, err := govalidator.ValidateStruct(h.Value)
		if err != nil || !res {
			ctx.PrintError(w, err)
			return
		}

		e.SaveDraft(h)

		ctx.PrintResult(w, ent)
	}
}

func (a *App) KindGetHandler(e *entity.Entity) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r)

		/*vars := mux.Vars(r)
		id := vars["id"]

		key, err := datastore.DecodeKey(id)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}*/

		/*h, err := k.Get(ctx, key)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}*/

		ctx.PrintResult(w, k)
	}
}

/*

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

		ctx.PrintResult(w, h.Output())
	}
}

func (a *App) KindDeleteHandler(k *kind.Kind) http.HandlerFunc {
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

		err = h.Delete(key)
		if err != nil {
			ctx.PrintError(w, err)
			return
		}

		ctx.PrintResult(w, h.Output())
	}
}
*/
