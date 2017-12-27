package cms

import (
	"github.com/PuerkitoBio/goquery"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type view map[string]interface{}

type Site struct {
	*SiteOptions
	/*Pages   *template.Template*/
	Index *goquery.Document

	Pages map[string]*Page
}
type SiteOptions struct {
	Bucket  string
	BaseDir string
	Routers []*Router
}
type Page struct {
	Document *goquery.Document
	Title    string
	Body     string
}
type Router struct {
	OutletSelector string
	Handler        Handler
}
type Handler map[string]string

func NewSite(opt *SiteOptions) *Site {
	// 1. Parse and render pages
	// 2. Save pages to bucket
	var site = &Site{
		SiteOptions: opt,
		Pages:       map[string]*Page{},
	}

	// Read index.html and query for HTML imports
	var indexPath = path.Join(opt.BaseDir, "index.html")
	site.Index = site.parseFile(indexPath)

	// Parse pages
	/*var err error
	site.Pages, err = template.New("").Funcs(htmlFuncMap).ParseGlob(path.Join(opt.BaseDir, "/pages/*.html"))
	if err != nil {
		panic(err)
	}*/

	matches, err := filepath.Glob(path.Join(opt.BaseDir, "/pages/*.html"))
	if err != nil {
		panic(err)
	}

	for _, pagePath := range matches {
		site.parsePage(pagePath)
	}

	/*for _, router := range site.Routers {
		for pageURL, pageName := range router.Handler {
			site.parsePage(pageURL, pageName)
		}
	}*/

	return site
}

func (s *Site) parsePage(filePath string) {
	var err error
	var pageName = path.Base(filePath)
	var spl = strings.Split(pageName, ".")
	pageName = strings.Join(spl[:len(spl)-1], ".")

	var pageDocument = s.parseFile(filePath)
	var page = &Page{
		Document: pageDocument,
	}

	title := pageDocument.Find("title").First()
	body := pageDocument.Find("body").First()

	page.Title = title.Text()
	page.Body, err = body.Html()
	if err != nil {
		panic(err)
	}

	s.Pages[pageName] = page
}

func (s *Site) parseFile(path string) *goquery.Document {
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
	s.parseHTMLComponents(document.Find("link[rel=import]"), body)
	/*rootHTML, err := body.Html()
	if err != nil {
		panic(err)
	}*/

	return document

	/*rootHTML, err = renderTemplate(rootHTML)
	if err != nil {
		panic(err)
	}
	body.SetHtml(rootHTML)

	documentHtml, err := document.Html()
	if err != nil {
		panic(err)
	}

	return documentHtml*/
}

func (s *Site) parseHTMLComponents(links *goquery.Selection, root *goquery.Selection) {
	// 1. Fetch all imports
	// 2. Find links inside the import
	// 3. Replace element content HTML with imported template content
	// 4. Return rendered root element
	links.Each(func(i int, linkNode *goquery.Selection) {
		if importHref, ok := linkNode.Attr("href"); ok {
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
						s.parseHTMLComponents(childLinks, element)
					})
				}
			})
		}
	})
}
