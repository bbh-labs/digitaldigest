package main

import (
	"io/ioutil"
	"log"
	"mime"
	"path"
	"strings"
)

type File struct {
	Name string
	Type string
}

func listFiles() []File {
	fileinfos, err := ioutil.ReadDir("content")
	if err != nil {
		log.Fatal(err)
	}

	var files []File
	for _, fileinfo := range fileinfos {
		var file File
		file.Name = "content/" + fileinfo.Name()

		var mimeType = mime.TypeByExtension(path.Ext(file.Name))
		if strings.Contains(mimeType, "image") {
			file.Type = "image"
		} else if strings.Contains(mimeType, "video") {
			file.Type = "video"
		}
		files = append(files, file)
	}

	return files
}
