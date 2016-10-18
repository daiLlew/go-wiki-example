package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
)

const txtFileExt = "%v.txt"
const viewPath = "/view/"
const editPath = "/edit/"

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

type Page struct {
	Title string
	Body  []byte
}

func main() {
	handlersMap := map[string]http.HandlerFunc{
		"/view/": makeHandler(viewHandler),
		"/edit/": makeHandler(editHandler),
		"/save/": makeHandler(saveHandler),
	}

	for name, handler := range handlersMap {
		fmt.Printf("\t[startup]Registering http.Handler: %v\n", name)
		http.Handle(name, handler)
	}

	http.ListenAndServe(":8080", nil)
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

func makeHandler(fn func(w http.ResponseWriter, r *http.Request, title string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Inside make handler")
		titleMatch := validPath.FindStringSubmatch(r.URL.Path)
		if titleMatch == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, titleMatch[2])
	}
}

/* View the requested page. */
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	fmt.Println("view handler")
	p, err := loadPage(title)
	if err != nil {
		fmt.Printf("[viewHandlerFunc] error loading file: %v\n", err.Error())
		http.Redirect(w, r, editPath + title, http.StatusFound)
		return
	}
	renderTemplate(p, "view.html", w)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		fmt.Printf("error loading file: %v\n", err.Error())
		p = &Page{Title: title}
	}
	renderTemplate(p, "edit.html", w)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err2 := p.save()

	if err2 != nil {
		fmt.Printf("\n[saveHandlerFunc] Something went wrong %v\n", err2.Error())
	}
	http.Redirect(w, r, viewPath + title, http.StatusFound)
}

func renderTemplate(p *Page, templateName string, w http.ResponseWriter) {
	err := templates.ExecuteTemplate(w, templateName, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}