package main

import (
	"log"
	"io"
	"io/ioutil"
	"html/template"
	"net/http"
	"net/textproto"
	"os"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

const MultipartMaxMemory = 4 * 1024 * 1024 // 4MB

type File struct {
	Filename string
	Type string
}

func listFiles() interface{} {
	fileinfos, err := ioutil.ReadDir("content")
	if err != nil {
		log.Fatal(err)
	}

	var files []File
	for _, fileinfo := range fileinfos {
		var file File
		file.Filename = "content/" + fileinfo.Name()
		files = append(files, file)
	}

	return files
}

func home(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "home", listFiles())
}

func admin(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "admin", listFiles())
}

func content(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		postContent(w, r)
	case "DELETE":
		deleteContent(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func postContent(w http.ResponseWriter, r *http.Request) {
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

	output, err := os.OpenFile("content/" + headers[0].Filename, os.O_CREATE, 0600)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer output.Close()

	file, err := headers[0].Open()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer file.Close()

	if _, err = io.Copy(output, file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteContent(w http.ResponseWriter, r *http.Request) {
	filepath := r.FormValue("filepath")

	if err := os.Remove(filepath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

var templates *template.Template

func isImage(mime textproto.MIMEHeader) bool {
	for _, v := range mime {
		for _, vv := range v {
			if strings.Contains(vv, "image") {
				return true
			}
		}
	}
	return false
}

func isVideo(mime textproto.MIMEHeader) bool {
	for _, v := range mime {
		for _, vv := range v {
			if strings.Contains(vv, "video") {
				return true
			}
		}
	}
	return false
}

func main() {
	if err := os.Mkdir("content", 0700); err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	templates = template.Must(template.ParseGlob("templates/*.html"))
	templates = templates.Funcs(map[string]interface{}{
		"isImage": isImage,
		"isVideo": isVideo,
	})

	router := mux.NewRouter()
	router.HandleFunc("/", home)
	router.HandleFunc("/admin", admin)
	router.HandleFunc("/content", content)

	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":8080")
}
