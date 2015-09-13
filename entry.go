package main

import (
	"io/ioutil"
	"log"
	"mime"
	"path"
	"os"
	"regexp"
	"strings"
)

type Entry struct {
	Name string
	Image string
	Video string
	IsYoutube bool
	IsVimeo bool
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
		isText := strings.HasPrefix(mimeType, "text")

		filename := "content/" + fileinfo.Name()
		if isImage {
			entry.Image = filename
		} else if isVideo {
			entry.Video = filename
		} else if isText && fileinfo.Size() <= 2000 {
			if entry.Video, err = readURLFromFile(filename); err != nil {
				continue
			}
			if entry.IsYoutube, err = regexp.Match(`.*youtube\.com\/.+`, []byte(entry.Video)); err != nil {
				continue
			} else if !entry.IsYoutube {
				if entry.IsVimeo, err = regexp.Match(`.*vimeo\.com\/.+`, []byte(entry.Video)); err != nil {
					continue
				}
			} else if !entry.IsYoutube && !entry.IsVimeo {
				continue
			}
		} else {
			continue
		}

		// It's the last file so just append the current entry
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

func readURLFromFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
