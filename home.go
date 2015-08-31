package main

import (
	"html/template"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Files []File
	}{
		Files: listFiles(),
	}

	if *refreshTemplates {
		if tmpls, err := template.New("t").ParseGlob("templates/*.html"); err == nil {
			templates = tmpls
		}
	}

	templates.ExecuteTemplate(w, "home", data)
}
