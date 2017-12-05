package cms

import (
	"net/http"
	"html/template"
	"google.golang.org/appengine"
)

var index *template.Template
var options map[string]interface{}
var LocalAssetsHost = "localhost:3000"

const CDNAssetsHost = "google.com"

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

    <title>CMS</title>

    <!-- Global stylesheet -->
    <link rel="stylesheet" href="//{{ .assetsHost }}/style.min.css">
    <!-- /global stylesheet -->

    <script src="https://cdn.jsdelivr.net/npm/navigo@6.0.0/lib/navigo.min.js"></script>
    <script src="https://unpkg.com/axios/dist/axios.min.js"></script>

    <script src="//{{ .assetsHost }}/components/custom.js"></script>
    <script src="//{{ .assetsHost }}/components/util.js"></script>
</head>
<body>

<div class="-app side"></div>

<script>
    // save entry component for global event handling; could also use customComponents.main
    var global = customComponents.init({
        baseURL: '//{{ .assetsHost }}/components/',
        main: document.body,
        imports: ['app']
    });
</script>
</body>
</html>
{{ end }}`)
	if err != nil {
		panic(err)
	}

	options = map[string]interface{}{}
	if appengine.IsDevAppServer() {
		options["assetsHost"] = LocalAssetsHost
	} else {
		options["assetsHost"] = CDNAssetsHost
	}
}

func editor() http.Handler {
	return http.Handler(http.HandlerFunc(renderEditor))
}

func renderEditor(w http.ResponseWriter, r *http.Request) {
	err := index.ExecuteTemplate(w, "editor", options)
	if err != nil {
		ctx := NewContext(r)
		ctx.PrintError(w, err, http.StatusInternalServerError)
	}
}
