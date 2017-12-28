package cms

import (
	"html/template"
	"net/http"
	"bytes"
)

func ParsePage(funcs template.FuncMap, templates ...string) (*template.Template, error) {
	return template.New("").Funcs(htmlFuncMap).Funcs(funcs).ParseFiles(templates...)
}

func RenderTemplate(w http.ResponseWriter, templ *template.Template, data interface{}) error {
	return templ.ExecuteTemplate(w, "index", data)
}

func (a *API) TemplateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := a.NewContext(r).WithBody()

	t, err := template.New("").Funcs(htmlFuncMap).Parse(`{{define "body"}}` + string(ctx.body.body) + "{{end}}")
	if err != nil {
		ctx.PrintError(w, err, http.StatusBadRequest)
		return
	}

	err = t.ExecuteTemplate(w, "body", data{"items": []string{"one", "two", "three"}})
	if err != nil {
		ctx.PrintError(w, err, http.StatusBadRequest)
	}
}

func renderTemplate(body string) (string, error) {
	t, err := template.New("").Funcs(htmlFuncMap).Parse(`{{define "body"}}` + body + "{{end}}")
	if err != nil {
		return "", err
	}

	var doc bytes.Buffer
	defer doc.Reset()
	err = t.ExecuteTemplate(&doc, "body", data{"items": []string{"one", "two", "three"}})
	if err != nil {
		return "", err
	}
	return doc.String(), nil
}

var htmlFuncMap = template.FuncMap{
	"valOfMap": func(x map[string]interface{}, key string) interface{} {
		return x[key]
	},
	"toHTML": func(s string) template.HTML {
		return template.HTML(s)
	},
	"toCSS": func(s string) template.CSS {
		return template.CSS(s)
	},
	"toJS": func(s string) template.JS {
		return template.JS(s)
	},
}
