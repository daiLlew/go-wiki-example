package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
)

const templatesPath = "%v.html"
const resourcesPath = "resources/%v"

var viewHandler WikiHandler
var editHandler WikiHandler
var saveHandler WikiHandler

//var errorHandler WikiHandler

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
	//errorHandler = WikiHandler{uri: "/error", handler: errorHandlerFunc

	http.ListenAndServe(":8080", nil)
	registerHandler(viewHandler, editHandler, saveHandler)
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
	filename := getFilename(r, viewHandler)
	p, err := loadPage(filename)
	if err != nil {
		fmt.Printf("[viewHandlerFunc] error loading file: %v\n", err.Error())
		http.Redirect(w, r, "/edit/"+filename, http.StatusFound)
		return
	}
	renderTemplate(p, viewHandler, w)
}

func editHandlerFunc(w http.ResponseWriter, r *http.Request) {
	p, err := loadPage(getFilename(r, editHandler))
	if err != nil {
		fmt.Printf("error loading file: %v\n", err.Error())
		p = &Page{Title: getFilename(r, editHandler)}
	}
	renderTemplate(p, editHandler, w)
}

func saveHandlerFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Inside saveHandlerFunc")
	title := getFilename(r, saveHandler)
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		fmt.Printf("\n[saveHandlerFunc] Something went wrong %v\n", err.Error())
	}
	http.Redirect(w, r, viewHandler.uri+title, http.StatusFound)
}

func renderTemplate(p *Page, handler WikiHandler, w http.ResponseWriter) {
	t, err := template.ParseFiles(handler.templatePath())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	execErr := t.Execute(w, p)
	if execErr != nil {
		http.Error(w, execErr.Error(), http.StatusInternalServerError)
		return
	}

}

func getFilename(r *http.Request, handler WikiHandler) string {
	return r.URL.Path[len(handler.uri):]
}

func (p *Page) save() error {
	filename := fmt.Sprintf(resourcesPath, p.Title+".txt")
	fmt.Printf("saving to %v", filename)
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := fmt.Sprintf(resourcesPath, title+".txt")
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func (h *WikiHandler) templatePath() string {
	templatePath := fmt.Sprintf(templatesPath, h.name())
	fmt.Printf("Template path for %v is %v\n", h.uri, templatePath)
	return templatePath
}

func (h *WikiHandler) name() string {
	return h.uri[1 : len(h.uri)-1]
}
