package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"errors"
	"regexp"
)

const txtFileExt = "%v.txt"
const htmlFileExt = "%v.html"

var viewHandler WikiHandler
var editHandler WikiHandler
var saveHandler WikiHandler
var errorHandler WikiHandler
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

type Page struct {
	Title string
	Body  []byte
}

type WikiHandler struct {
	uri     string
	handler http.HandlerFunc
}

func main() {
	viewHandler = WikiHandler{uri: "/view/", handler: viewHandlerFunc}
	editHandler = WikiHandler{uri: "/edit/", handler: editHandlerFunc}
	saveHandler = WikiHandler{uri: "/save/", handler: saveHandlerFunc}
	errorHandler = WikiHandler{uri: "/error", handler: errorHandlerFunc}

	registerHandler(viewHandler, editHandler, saveHandler, errorHandler)
	http.ListenAndServe(":8080", nil)
}

/* Register the Wiki handlers */
func registerHandler(wikiHandlers ...WikiHandler) {
	for _, handler := range wikiHandlers {
		fmt.Printf("[startup] Registering http.Handler: %v\n", handler.name())
		http.HandleFunc(handler.uri, handler.handler)
	}
}

/* View the requested page. */
func viewHandlerFunc(w http.ResponseWriter, r *http.Request) {
	filename, err := getTitle(w, r)
	if err != nil {
		return
	}
	p, err := loadPage(filename)
	if err != nil {
		fmt.Printf("[viewHandlerFunc] error loading file: %v\n", err.Error())
		http.Redirect(w, r, editHandler.uri + filename, http.StatusFound)
		return
	}
	renderTemplate(p, viewHandler, w)
}

func editHandlerFunc(w http.ResponseWriter, r *http.Request) {
	filename, err := getTitle(w, r)
	if err != nil {
		return
	}
	p, err := loadPage(filename)
	if err != nil {
		fmt.Printf("error loading file: %v\n", err.Error())
		p = &Page{Title: filename}
	}
	renderTemplate(p, editHandler, w)
}

func saveHandlerFunc(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err2 := p.save()

	if err2 != nil {
		fmt.Printf("\n[saveHandlerFunc] Something went wrong %v\n", err2.Error())
	}
	http.Redirect(w, r, viewHandler.uri + title, http.StatusFound)
}

func errorHandlerFunc(w http.ResponseWriter, r *http.Request) {
	err := errors.New("This aint working no more.")
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}

func renderTemplate(p *Page, handler WikiHandler, w http.ResponseWriter) {
	err := templates.ExecuteTemplate(w, handler.templateName(), p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (p *Page) save() error {
	filename := fmt.Sprintf(txtFileExt, p.Title)
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := fmt.Sprintf(txtFileExt, title)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func (h *WikiHandler) templateName() string {
	templatePath := fmt.Sprintf(htmlFileExt, h.name())
	fmt.Printf("Template path for %v is %v\n", h.uri, templatePath)
	return templatePath
}

func (h *WikiHandler) name() string {
	if lastChar := string(h.uri[len(h.uri) - 1]); lastChar == "/" {
		return h.uri[1 : len(h.uri) - 1]
	}
	return h.uri[1:]
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // The title is the second subexpression.
}