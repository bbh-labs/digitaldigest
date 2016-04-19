package main

import (
	"net/http"

	"github.com/bbh-labs/digitaldigest/session"
)

func logout(w http.ResponseWriter, r *http.Request) {
	session.Clear(w, r)
}
