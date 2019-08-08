package main

import (
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

const maxUploadSize = 4 * 1024 * 1024 * 1024 // 4GB

func main() {

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/sas", sasHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/progress", progressHandler)
	http.HandleFunc("/", homeHandler)
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {

	blobs, err := listBlobs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lp := filepath.Join("templates", "home.html")

	tmpl, err := template.ParseFiles(lp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tmpl.ExecuteTemplate(w, "home", struct {
		Blobs []azblob.BlobItem
	}{
		blobs,
	})
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
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

	err = uploadToStorageBlob(handler.Filename, file, handler.Size)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	http.Redirect(w, r, "/", 302)
}

func sasHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.FormValue("filename")

	url := getSAS(filename)

	w.Write([]byte(url))
}

func progressHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query()["file"][0]

	if val, ok := progress.Load(fileName); ok {
		w.Write([]byte(strconv.Itoa(val.(int))))
		return
	}

	http.Error(w, "File not found", 404)
}
