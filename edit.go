package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path"
	"os"
	"strings"

	"github.com/bbhmakerlab/digitaldigest/session"
	"github.com/bbhmakerlab/digitaldigest/store"
)

const (
	MultipartMaxMemory = 4 * 1024 * 1024        // 4MB
	totalDiskSpace     = 1 * 1024 * 1024 * 1024 // 1GB
)

func edit(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getContent(w, r)
	case "POST":
		postContent(w, r)
	case "DELETE":
		deleteContent(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func getContent(w http.ResponseWriter, r *http.Request) {
	data := struct {
		UsedDiskSpacePercentage string
		Entries                 []store.Entry
		IsLoggedIn              bool
	}{
		UsedDiskSpacePercentage: fmt.Sprintf("%.1f", float64(usedDiskSpace())/float64(totalDiskSpace)*100),
		Entries: store.GetEntries(-1),
		IsLoggedIn: session.GetEmail(r) != "",
	}

	templates.ExecuteTemplate(w, "edit", data)
}

func postContent(w http.ResponseWriter, r *http.Request) {
	if session.GetEmail(r) == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	mimeType := ""

	// User uploaded URL
	url := r.FormValue("url")
	if url != "" {
		if strings.Contains(url, "youtube.com/") {
			mimeType = "video/youtube"
		} else if strings.Contains(url, "vimeo.com/") {
			mimeType = "video/vimeo"
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		store.InsertEntry(store.Entry{URL: url, MIMEType: mimeType})
		return
	}

	// User uploaded file
	if err := r.ParseMultipartForm(MultipartMaxMemory); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	headers := r.MultipartForm.File["file"]
	if len(headers) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("no files were uploaded")
		return
	}

	for _, header := range headers {
		file, err := header.Open()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		defer file.Close()

		// Create the 'content' directory if doesn't exist
		if _, err := os.Stat("content"); err != nil {
			if os.IsNotExist(err) {
				if err = os.Mkdir("content", 0700); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Println(err)
					return
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println(err)
				return
			}
		}

		url := "content/" + header.Filename
		output, err := os.OpenFile(url, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			os.Remove(url)
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		defer output.Close()

		if _, err = io.Copy(output, file); err != nil {
			os.Remove(url)
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		mimeType = mime.TypeByExtension(path.Base(header.Filename))
		store.InsertEntry(store.Entry{URL: url, MIMEType: mimeType})
	}

	http.Redirect(w, r, "/edit", http.StatusFound)
}

func deleteContent(w http.ResponseWriter, r *http.Request) {
	if session.GetEmail(r) == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	typ := r.FormValue("type")
	if typ == "all" {
		store.DeleteCurrentEntries()
		return
	}

	file := r.FormValue("file")

	if err := os.Remove(file); err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("File " + file + " was not found!"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func usedDiskSpace() int64 {
	fileinfos, err := ioutil.ReadDir("content")
	if err != nil {
		log.Fatal(err)
		return -1
	}

	var total int64
	for _, fileinfo := range fileinfos {
		total += fileinfo.Size()
	}

	return total
}
