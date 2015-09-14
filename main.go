package main

import (
	"flag"
	"html/template"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

var templates *template.Template
var refreshTemplates = flag.Bool("refresh", false, "Refresh templates on every page load")

func main() {
	flag.Parse()

	templates = template.Must(template.New("t").ParseGlob("templates/*.html"))

	router := mux.NewRouter()
	router.HandleFunc("/", home)
	router.HandleFunc("/edit", edit)
	router.HandleFunc("/edit/image", editImage)
	router.HandleFunc("/login", login)
	router.HandleFunc("/logout", logout)
	router.Handle("/ws/home", homeWS)
	router.Handle("/ws/edit", editWS)
	go func() {
		wsHub.run()
	}()

	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":8080")
}
