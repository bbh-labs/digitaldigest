package main

import (
	"html/template"
	"log"
	"os"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

var templates *template.Template

func main() {
	if err := os.Mkdir("content", 0700); err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	templates = template.Must(template.New("t").ParseGlob("templates/*.html"))

	router := mux.NewRouter()
	router.HandleFunc("/", home)
	router.HandleFunc("/edit", edit)
	router.HandleFunc("/login", login)
	router.HandleFunc("/logout", logout)
	router.Handle("/ws", ws)
	go func() {
		ws.h.run()
	}()

	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":8080")
}
