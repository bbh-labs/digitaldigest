package main

import (
	"net/http"

	"github.com/bbhasiapacific/digitaldigest/session"
)

func logout(w http.ResponseWriter, r *http.Request) {
	session.Clear(w, r)
}
