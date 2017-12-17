package cms

import (
	"github.com/PuerkitoBio/goquery"
	"os"
	"path"
)

func getPath(base string, file string) (string, string) {
	if file[:2] == ".." {
		file = file[1:]
	}

	newPath := path.Join(base, path.Clean(file))

	return path.Dir(newPath), newPath
}

type view map[string]interface{}

func NewPageIndex(basePath string, indexPath string) string {
	basePath = path.Clean(basePath)

	basePath, indexPath = getPath(basePath, indexPath)

	indexFile, err := os.Open(indexPath)
	if err != nil {
		panic(err)
	}
	document, err := goquery.NewDocumentFromReader(indexFile)
	indexFile.Close()
	if err != nil {
		panic(err)
	}

	// find all go-root instances
	document.Find("go-root").Each(func(i int, root *goquery.Selection) {
		var _, isRendered = root.Attr("rdy")

		// fetch settings
		/*if settingsPath, ok := root.Attr("settings"); ok {

		}*/

		// init go-root rendering
		if !isRendered {
			parseHTMLComponents(basePath, document.Find("head link[rel=import]"), root)
			var rootHTML, err = root.Html()
			if err != nil {
				panic(err)
			}
			rootHTML, err = renderTemplate(rootHTML)
			if err != nil {
				panic(err)
			}
			root.SetHtml(rootHTML)
			root.SetAttr("rdy", "")
		}
	})

	documentHtml, err := document.Html()
	if err != nil {
		panic(err)
	}

	return documentHtml
}

func parseHTMLComponents(basePath string, links *goquery.Selection, root *goquery.Selection) {
	// 1. Fetch all imports
	// 2. Find links inside the import
	// 3. Replace element content HTML with imported template content
	// 4. Return rendered root element
	links.Each(func(i int, linkNode *goquery.Selection) {
		if importHref, ok := linkNode.Attr("href"); ok {
			// read imported files
			newBase, importHref := getPath(basePath, importHref)
			importFile, err := os.Open(importHref)
			if err != nil {
				panic(err)
			}
			link, err := goquery.NewDocumentFromReader(importFile)
			importFile.Close()
			if err != nil {
				panic(err)
			}

			// todo: find link[rel=import] inside the import
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
						parseHTMLComponents(newBase, childLinks, element)
					})
				}
			})
		}
	})
}
