package main

import (
	"flag"
	"net/http"
	"strings"

	"github.com/bbhmakerlab/digitaldigest/session"
	"github.com/google/google-api-go-client/plus/v1"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var redirectURL = flag.String("url", "http://localhost:8080", "redirect URL for OAuth2")

var conf = &oauth2.Config{
	ClientID:     "275859936684-90o26gr4hdbr4jgvdjobuath4qhq90fc.apps.googleusercontent.com",
	ClientSecret: "G-rp5gffbNDQgMgAhMxT5I7m",
	RedirectURL:  *redirectURL,
	Endpoint:     google.Endpoint,
}

func login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if session.GetEmail(r) != "" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	case "POST":
		authCode := r.FormValue("authCode")
		conf.RedirectURL = *redirectURL

		tok, err := conf.Exchange(oauth2.NoContext, authCode)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		client := conf.Client(oauth2.NoContext, tok)
		service, err := plus.New(client)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		call := service.People.Get("me")
		person, err := call.Do()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		email := ""
		for _, em := range person.Emails {
			if em.Type == "account" {
				email = em.Value
				break
			}
		}

		if !strings.HasSuffix(email, "@bartleboglehegarty.com") {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		session.Set(w, r, email)
		w.WriteHeader(http.StatusOK)
	}
}
