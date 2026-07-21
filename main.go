package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/syumai/spidermonkey-api/internal/server"
)

// The embed directive is package-relative, so it (and the "public"
// directory it captures) has to stay at the module root; internal/server
// receives the resulting fs.FS instead of embedding it itself.
//
//go:embed public
var publicFS embed.FS

func main() {
	sub, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(http.ListenAndServe(":"+port, server.New(sub)))
}
