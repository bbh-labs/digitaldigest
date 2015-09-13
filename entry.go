package main

import (
	"io/ioutil"
	"log"
	"mime"
	"path"
	"strings"
)

type Entry struct {
	Name string
	Image string
	Video string
}

func countImages(entries []Entry) int {
	cnt := 0

	for _, entry := range entries {
		if entry.Image != "" {
			cnt++
		}
	}

	return cnt
}

func countVideos(entries []Entry) int {
	cnt := 0

	for _, entry := range entries {
		if entry.Video != "" {
			cnt++
		}
	}

	return cnt
}

func listEntries() []Entry {
	fileinfos, err := ioutil.ReadDir("content")
	if err != nil {
		log.Fatal(err)
	}

	var entries []Entry
	var entry Entry

	for i, fileinfo := range fileinfos {
		name := filenameWithoutExtension(fileinfo.Name())
		if entry.Name != "" && entry.Name != name {
			entries = append(entries, entry)
			entry = Entry{}
		}
		entry.Name = name

		mimeType := mime.TypeByExtension(path.Ext(fileinfo.Name()))
		isImage := strings.HasPrefix(mimeType, "image")
		isVideo := strings.HasPrefix(mimeType, "video")

		if isImage {
			entry.Image = "content/" + fileinfo.Name()
		} else if isVideo {
			entry.Video = "content/" + fileinfo.Name()
		}

		if i == len(fileinfos) - 1 {
			entries = append(entries, entry)
		}
	}

	return entries
}

func filenameWithoutExtension(filename string) string {
	ext := path.Ext(path.Base(filename))
	return filename[:len(filename) - len(ext)]
}
