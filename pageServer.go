package cms

import (
	"github.com/PuerkitoBio/goquery"
	"os"
	"path"
	"path/filepath"
	"strings"
	"net/http"
	"github.com/gorilla/mux"
	"html/template"
	"errors"
	"bytes"
	"google.golang.org/appengine/datastore"
)

type data map[string]interface{}

type Site struct {
	*SiteOptions
	/*Pages   *template.Template*/
	index *goquery.Document

	pages   map[string]*Page
	handler Handler
}
type SiteOptions struct {
	API     *API
	BaseDir string
	Bucket  string
}
type Page struct {
	Title string
	Body  template.HTML
	Data  data

	parsed *template.Template
	links  *goquery.Selection
}
type Handler map[string]*Page

var funcs = template.FuncMap{
	"get": func(ctx Context, id string) map[string]interface{} {
		key, err := datastore.DecodeKey(id)
		if err != nil {
			panic(err)
		}

		if e, ok := ctx.api.entities[key.Kind()]; ok {
			var dataHolder = e.New(ctx)
			dataHolder.key = key

			err = datastore.Get(ctx.Context, key, dataHolder)
			if err != nil {
				panic(err)
			}

			return dataHolder.Output(ctx, true)
		} else {
			panic(errors.New("entity " + key.Kind() + " doesn't exist"))
		}

		return nil
	},
	"setTitle": func(dest data, value string) interface{} {
		dest["page"].(*Page).Title = value
		return nil
	},
	"setValue": func(dest data, name string, value interface{}) interface{} {
		dest["page"].(*Page).Data[name] = value
		return nil
	},
	"getValue": func(dest data, name string) interface{} {
		return dest["page"].(*Page).Data[name]
	},
}

func NewSite(opt *SiteOptions) *Site {
	var site = &Site{
		SiteOptions: opt,
		pages:       map[string]*Page{},
	}

	// 1. Build index
	var indexPath = path.Join(opt.BaseDir, "index.html")
	site.index = site.buildFile(indexPath)

	// 2. Build pages
	matches, err := filepath.Glob(path.Join(opt.BaseDir, "/pages/*.html"))
	if err != nil {
		panic(err)
	}
	for _, pagePath := range matches {
		site.buildPage(pagePath)
	}

	return site
}

func (s *Site) Handler(handler Handler) http.Handler {
	s.handler = handler

	var indexImports = map[string]bool{}
	s.index.Find("link[rel=import]").Each(func(i int, selection *goquery.Selection) {
		if href, ok := selection.Attr("href"); ok {
			indexImports[path.Clean(href)] = true
		}
	})

	// 3. Single page rendering function
	var handlingFunc = func(page *Page) http.HandlerFunc {
		index := s.index.Clone()

		indexHead := index.Find("head")
		page.links.Each(func(i int, selection *goquery.Selection) {
			if href, ok := selection.Attr("href"); ok {
				var cleanHref = path.Clean(href)
				if _, ok := indexImports[cleanHref]; !ok {
					indexHead.AppendSelection(selection)
					indexImports[cleanHref] = true
				}
			}

		})

		// 2. Parse index html template
		indexHtml, err := index.Html()
		if err != nil {
			panic(err)
		}
		parsedIndex, err := template.New("").Funcs(funcs).Parse(indexHtml)
		if err != nil {
			panic(err)
		}

		return func(w http.ResponseWriter, r *http.Request) {
			var ctx Context
			if s.API != nil {
				ctx = s.API.newRendererContext(r)
			}

			var passingData = data{
				"context": ctx,
				"request": data{
					"vars": mux.Vars(r),
				},
				"page": page,
			}

			buf := new(bytes.Buffer)
			err := page.parsed.Execute(buf, passingData)
			if err != nil {
				ctx.PrintError(w, err, http.StatusInternalServerError)
				return
			}

			page.Body = template.HTML(buf.String())

			parsedIndex.Execute(w, passingData)
		}
	}

	// 4. Set-up handlers
	var router = mux.NewRouter()
	for pagePath, page := range s.handler {
		router.HandleFunc(pagePath, handlingFunc(page))
	}

	return router
}

func (s *Site) Page(name string) *Page {
	if page, ok := s.pages[name]; ok {
		return page
	}
	panic(errors.New("page " + name + " doesn't exist"))
	return nil
}

func (s *Site) buildPage(filePath string) *Page {
	var pageName = path.Base(filePath)
	var spl = strings.Split(pageName, ".")
	pageName = strings.Join(spl[:len(spl)-1], ".")

	var pageDocument = s.buildFile(filePath)

	var page = &Page{
		Data: data{},
	}

	// get links
	var links = pageDocument.Find("link[rel=import]")
	page.links = links.Clone()

	// remove links
	links.Remove()

	var html, err = pageDocument.Html()
	if err != nil {
		panic(err)
	}

	page.parsed, err = template.New("").Funcs(funcs).Parse(html)
	if err != nil {
		panic(err)
	}

	s.pages[pageName] = page

	return page
}

func (s *Site) buildFile(path string) *goquery.Document {
	indexFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	document, err := goquery.NewDocumentFromReader(indexFile)
	indexFile.Close()
	if err != nil {
		panic(err)
	}

	var body = document.Find("body")
	s.buildImports(document.Find("link[rel=import]"), body)

	return document
}

func (s *Site) buildImports(links *goquery.Selection, root *goquery.Selection) {
	// 1. Fetch all imports
	// 2. Find links inside the import
	// 3. Replace element content HTML with imported template content
	// 4. Return rendered root element
	links.Each(func(i int, linkNode *goquery.Selection) {
		_, goRender := linkNode.Attr("go")
		if importHref, ok := linkNode.Attr("href"); ok && goRender {
			linkNode.RemoveAttr("go")

			// read imported files
			importHref := path.Join(s.BaseDir, importHref)
			importFile, err := os.Open(importHref)
			if err != nil {
				panic(err)
			}
			link, err := goquery.NewDocumentFromReader(importFile)
			importFile.Close()
			if err != nil {
				panic(err)
			}

			var childLinks = link.Find("link[rel=import]")

			// find templates inside the import
			var templates = link.Find("template")
			templates.Each(func(i int, template *goquery.Selection) {
				// use only templates with defined id's
				if id, ok := template.Attr("id"); ok {
					// query root element for elements with names matching the template's id
					root.Find(id).Each(func(i int, element *goquery.Selection) {
						var templateHTML, err = template.Html()
						if err != nil {
							panic(err)
						}
						element.SetHtml(templateHTML)
						s.buildImports(childLinks, element)
					})
				}
			})

			// remove link import from root if it doesn't contain script tag
			if link.Find("script").Length() == 0 {
				linkNode.Remove()
			}
		}
	})
}
