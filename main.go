package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/Azure/azure-storage-blob-go/azblob"
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
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}

	//ioutil.WriteFile(fmt.Sprintf("uploads/%s", handler.Filename), fileBytes, 0644)
	err = uploadToStorageBlob(handler.Filename, fileBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	http.Redirect(w, r, "/", 302)
}
