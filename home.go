package main

import (
	"html/template"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	entries := listEntries()

	data := struct {
		Entries []Entry
		NumImages int
		NumVideos int
	}{
		Entries: entries,
		NumImages: countImages(entries),
		NumVideos: countVideos(entries),
	}

	if *refreshTemplates {
		if tmpls, err := template.New("t").ParseGlob("templates/*.html"); err == nil {
			templates = tmpls
		}
	}

	templates.ExecuteTemplate(w, "home", data)
}
