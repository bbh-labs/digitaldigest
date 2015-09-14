package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"path"
	"os"
	"regexp"
	"strings"

	"github.com/bbhmakerlab/digitaldigest/session"
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
		switch r.FormValue("type") {
		case "files":
			uploadContent(w, r)
		case "url":
			uploadLink(w, r)
		}
	case "DELETE":
		deleteContent(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func editImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		return
	}

	if session.GetEmail(r) == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	headers, err := prepareMultipartForm(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if headers == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	deletePreviousFile(name)

	saveMediaFile(headers[0], name)
}

func getContent(w http.ResponseWriter, r *http.Request) {
	data := struct {
		UsedDiskSpacePercentage string
		Entries                 []Entry
		IsLoggedIn              bool
	}{
		UsedDiskSpacePercentage: fmt.Sprintf("%.1f", float64(usedDiskSpace())/float64(totalDiskSpace)*100),
		Entries: listEntries(),
		IsLoggedIn: session.GetEmail(r) != "",
	}

	if *refreshTemplates {
		if tmpls, err := template.New("t").ParseGlob("templates/*.html"); err == nil {
			templates = tmpls
		}
	}

	w.WriteHeader(http.StatusOK)
	templates.ExecuteTemplate(w, "edit", data)
}

func uploadContent(w http.ResponseWriter, r *http.Request) {
	if session.GetEmail(r) == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	headers, err := prepareMultipartForm(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if headers == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, header := range headers {
		if err = saveMediaFile(header, ""); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	http.Redirect(w, r, "/edit", http.StatusFound)
}

func uploadLink(w http.ResponseWriter, r *http.Request) {
	if session.GetEmail(r) == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	name := r.FormValue("name")

	// Check if there's an entry that has same name
	entries := listEntries()
	for _, entry := range entries {
		if name == entry.Name {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	url := r.FormValue("url")

	// Check if the url is valid
	matched, err := regexp.Match(`(.*youtube\.com\/.+)|(.*vimeo\.com\/.+)`, []byte(url))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !matched {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create the 'content' directory if doesn't exist
	if err = prepareContentDirectory(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Create URL file
	filepath := "content/" + name + ".txt"
	output, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		os.Remove(filepath)
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer output.Close()

	// Write URL to file
	output.Write([]byte(url))

	http.Redirect(w, r, "/edit", http.StatusFound)
}

func deleteContent(w http.ResponseWriter, r *http.Request) {
	if session.GetEmail(r) == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	typ := r.FormValue("type")
	if typ == "all" {
		if err := deleteCurrentFiles(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
		}
		return
	}

	fileinfos, err := ioutil.ReadDir("content")
	if err != nil {
		log.Fatal(err)
	}

	name := r.FormValue("name")
	for _, fileinfo := range fileinfos {
		filename := fileinfo.Name()
		matching := name == filenameWithoutExtension(filename)
		if matching {
			os.Remove("content/" + filename)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func deletePreviousFile(name string) error {
	fileinfos, err := ioutil.ReadDir("content")
	if err != nil {
		log.Fatal(err)
	}

	for _, fileinfo := range fileinfos {
		filename := fileinfo.Name()
		if n := filenameWithoutExtension(filename); name == n {
			if err := os.Remove("content/" + filename); err != nil {
				return err
			}
			log.Println("Deleted file:", filename)
			return nil
		}
	}

	return nil
}

func deleteCurrentFiles() error {
	fileinfos, err := ioutil.ReadDir("content")
	if err != nil {
		log.Fatal(err)
	}

	for _, fileinfo := range fileinfos {
		if err := os.Remove("content/" + fileinfo.Name()); err != nil {
			return err
		}
	}

	return nil
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

func prepareMultipartForm(r *http.Request) ([]*multipart.FileHeader, error) {
	if err := r.ParseMultipartForm(MultipartMaxMemory); err != nil {
		log.Println(err)
		return nil, err
	}

	headers := r.MultipartForm.File["file"]
	if len(headers) == 0 {
		log.Println("no files were uploaded")
		return nil, nil
	}

	return headers, nil
}

func prepareContentDirectory() error {
	if _, err := os.Stat("content"); err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir("content", 0755); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func saveMediaFile(header *multipart.FileHeader, name string) error {
	ext := path.Ext(header.Filename)
	mimeType := mime.TypeByExtension(ext)
	isImage := strings.HasPrefix(mimeType, "image")
	isVideo := strings.HasPrefix(mimeType, "video")
	if (isImage || isVideo) == false {
		return nil
	}

	file, err := header.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	// Create the 'content' directory if doesn't exist
	if err = prepareContentDirectory(); err != nil {
		return err
	}

	var filepath string
	if name != "" {
		filepath = "content/" + name + ext
	} else {
		filepath = "content/" + header.Filename
	}
	output, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		os.Remove(filepath)
		return err
	}
	defer output.Close()

	if _, err = io.Copy(output, file); err != nil {
		os.Remove(filepath)
		return err
	}

	return nil
}
