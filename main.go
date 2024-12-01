package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"log"
	"net/http"
	"strconv"

	"github.com/disintegration/imaging"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	http.HandleFunc("/compress", compressImageHandler)
	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func compressImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve the file
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Unable to retrieve the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read quality from query parameters (default is 75)
	quality := 75
	if q := r.URL.Query().Get("quality"); q != "" {
		quality, err = strconv.Atoi(q)
		if err != nil || quality < 1 || quality > 100 {
			http.Error(w, "Quality must be an integer between 1 and 100", http.StatusBadRequest)
			return
		}
	}

	// Decode the uploaded image
	img, err := imaging.Decode(file)
	if err != nil {
		http.Error(w, "Unsupported image format", http.StatusUnsupportedMediaType)
		return
	}

	// Compress the image
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	if err != nil {
		http.Error(w, "Failed to compress the image", http.StatusInternalServerError)
		return
	}

	// Set headers and return the compressed image
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=compressed_%s", header.Filename))
	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}
