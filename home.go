package main

import (
	"net/http"

	"github.com/bbhmakerlab/digitaldigest/store"
)

func home(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Entries []store.Entry
	}{
		Entries: store.GetEntries(-1),
	}

	templates.ExecuteTemplate(w, "home", data)
}
