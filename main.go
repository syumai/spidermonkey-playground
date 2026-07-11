package main

import (
	"embed"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
)

//go:embed public
var publicFS embed.FS

func main() {
	http.HandleFunc("POST /api/eval", func(w http.ResponseWriter, req *http.Request) {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result, err := Eval(string(b))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(result))
	})

	sub, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", http.FileServerFS(sub))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	http.ListenAndServe(":"+port, nil)
}
