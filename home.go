package main

import (
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Files []File
	}{
		Files: listFiles(),
	}

	templates.ExecuteTemplate(w, "home", data)
}
