package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const maxUploadSize = 2 * 1024 * 1024 * 1024 // 2 mb

func main() {

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/", homeHandler)
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/home.html")
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	//r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "FILE_TOO_BIG", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}
	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}

	ioutil.WriteFile(fmt.Sprintf("uploads/%s", handler.Filename), fileBytes, 0644)

	w.Write([]byte("done"))
}
