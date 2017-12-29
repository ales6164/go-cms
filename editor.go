package cms

import (
	"net/http"
	"html/template"
	"google.golang.org/appengine"

	"github.com/asaskevich/govalidator"
	url2 "net/url"
	EngineUser "google.golang.org/appengine/user"
	"google.golang.org/appengine/datastore"
	"github.com/gorilla/sessions"
	"github.com/gorilla/securecookie"
	"golang.org/x/net/context"
)

var index *template.Template
var LocalAssetsHost = "localhost:3000"

var CDNAssetsHost = "google.com"

var store = sessions.NewCookieStore(securecookie.GenerateRandomKey(64))
var tokenKey = securecookie.GenerateRandomKey(64)

func (a *API) editor() http.Handler {
	return http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := a.NewContext(r)

		var err error
		var options = map[string]interface{}{}
		if appengine.IsDevAppServer() {
			options["assetsHost"] = LocalAssetsHost
		} else {
			options["assetsHost"] = CDNAssetsHost
		}

		// Own Auth
		session, err := store.Get(r, "id-token")

		if err == nil && session.Values["username"] != nil {

			// Already logged in
			if _, ok := r.URL.Query()["logout"]; ok {
				delete(session.Values, "username")
				delete(session.Values, "token")
				delete(session.Values, "count")
				session.Save(r, w)

				googleLogoutURL, err := EngineUser.LogoutURL(ctx.Context, r.URL.Path)
				if err == nil {
					http.Redirect(w, r, googleLogoutURL, http.StatusSeeOther)
				} else {
					http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				}
				return
			} else {
				session.Values["count"] = session.Values["count"].(int) + 1
			}
		} else {
			// Not logged in
			var ua *UserAccount
			var userKey *datastore.Key

			// Google Auth
			googleUser := EngineUser.Current(ctx.Context)
			if googleUser == nil {
				url, _ := EngineUser.LoginURL(ctx.Context, r.URL.Path)
				options["googleLogin"] = url
			} else {
				// Authenticated with Google
				// 1. Try fetching user from datastore OR creating Google user entry in datastore
				// 2. Create a session

				err = datastore.RunInTransaction(ctx.Context, func(tc context.Context) error {
					userKey, ua, err = getUser(tc, googleUser.Email)
					if err != nil {

						// User doesn't exist - try adding
						if err == datastore.ErrNoSuchEntity {

							// user doesn't exist - create new one
							ua = &UserAccount{
								Email:     googleUser.Email,
								UserGroup: a.options.DefaultUserGroup,
							}

							userKey, err = datastore.Put(tc, userKey, ua)

							return err
						}
					}
					return err

				}, nil)
				if err != nil {
					ctx.PrintError(w, err, http.StatusInternalServerError)
					return
				}

			}

			if len(ua.Email) > 0 {
				ctx, err := ctx.newToken(userKey, "admin", tokenKey)
				if err != nil {
					ctx.PrintError(w, err, http.StatusInternalServerError)
					return
				}

				// logged in!
				onUserLoggedIn(ctx, w, session)
				return
			} else {
				var isMethodPost = r.Method == http.MethodPost
				var username = r.PostFormValue("username") // always email
				var password = r.PostFormValue("password")

				if isMethodPost && len(username) > 0 {
					if govalidator.IsEmail(username) && len(password) > 0 {
						ctx, err := a.login(ctx, username, password)
						if err == nil {
							// logged in!
							onUserLoggedIn(ctx, w, session)
							return
						} else {
							options["loginFormError"] = err.Error()
						}
					} else {
						options["loginFormError"] = "invalid email or password"
					}
				} else {
					options["loginFormError"] = ""
				}

				// display login dialog
				err = index.ExecuteTemplate(w, "login", options)
				if err != nil {
					ctx.PrintError(w, err, http.StatusInternalServerError)
				}
				return
			}
		}

		options["token"] = session.Values["token"]
		options["user"] = session.Values["username"]
		options["handledEntities"] = a.handledEntities

		w.Header().Set("Content-type", "text/html; charset=utf-8")
		err = index.ExecuteTemplate(w, "editor", options)
		if err != nil {
			ctx.PrintError(w, err, http.StatusInternalServerError)
		}
	}))
}

func onUserLoggedIn(ctx Context, w http.ResponseWriter, session *sessions.Session) {
	// logged in!
	var continueURL = ctx.r.FormValue("continue")
	if len(continueURL) == 0 {
		continueURL = ctx.r.URL.Path
	} else {
		url1, err := url2.Parse(continueURL)
		if err == nil {
			continueURL = url1.Path
		} else {
			continueURL = ctx.r.URL.Path
		}
	}
	session.Values["count"] = 1
	session.Values["token"] = ctx.token
	session.Values["username"] = ctx.User().StringID()
	err := session.Save(ctx.r, w)
	if err != nil {
		ctx.PrintError(w, err, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, ctx.r, continueURL, http.StatusSeeOther)
}

func init() {
	var err error

	index, err = template.New("").Parse(`{{ define "editor" }}
<!DOCTYPE html>
<html lang="en">
<head>
    <base href="/">
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">

    <title>Sample Project</title>

    <!-- Global stylesheet -->
    <link rel="stylesheet" href="//{{ .assetsHost }}/style.min.css">
    <!-- /global stylesheet -->

    <script src="https://cdn.jsdelivr.net/npm/navigo@6.0.0/lib/navigo.min.js"></script>
    <script src="https://unpkg.com/axios/dist/axios.min.js"></script>

    <script src="//{{ .assetsHost }}/components/custom.js"></script>
    <script src="//{{ .assetsHost }}/components/util.js"></script>

	<script id="handledEntities" type="application/json">{{ .handledEntities }}</script>
</head>
<body>

<div class="-app side" data-token='{{ .token }}' data-user="{{ .user }}"></div>

<script>
    window["AssetsHost"] = "//" + {{ .assetsHost }};
    window["AppInstance"] = customComponents.init({
        baseURL: '//{{ .assetsHost }}/components/',
        main: document.body,
        imports: ['app']
    });
</script>
</body>
</html>
{{ end }}
{{ define "login" }}
<!DOCTYPE html>
<html lang="en">
<head>
    <base href="/">
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">

    <title>Sample Project</title>

    <!-- Global stylesheet -->
    <link rel="stylesheet" href="//{{ .assetsHost }}/style.min.css">
    <!-- /global stylesheet -->

    <script src="https://cdn.jsdelivr.net/npm/navigo@6.0.0/lib/navigo.min.js"></script>
    <script src="https://unpkg.com/axios/dist/axios.min.js"></script>

    <script src="//{{ .assetsHost }}/components/custom.js"></script>
    <script src="//{{ .assetsHost }}/components/util.js"></script>
</head>
<body>

<form method="post" action="/">
<label><span>Username</span><input required name="username" type="text"></label>
<label><span>Password</span><input required name="password" type="password"></label>
{{ with .loginFormError }} <span>{{ . }}</span> {{ end }}
<input type="submit" value="Login">
</form>

{{ with .googleLogin }}<a href="{{ . }}">Login with Google</a>{{ end }}
</body>
</html>
{{ end }}`)
	if err != nil {
		panic(err)
	}
}
