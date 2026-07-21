// Package server wires the playground's HTTP surface: POST /api/eval runs
// JavaScript through internal/jsruntime, and everything else serves the
// static playground UI.
package server

import (
	"context"
	"io"
	"io/fs"
	"net/http"
	"time"

	"github.com/syumai/spidermonkey-api/internal/jsruntime"
)

// evalTimeout bounds a single /api/eval request. It is generous enough to
// cover importing and evaluating "hono" (observed well under 100ms) with
// headroom for a slower host, while still keeping a runaway script (e.g.
// while(true){}) from holding a request open indefinitely.
const evalTimeout = 3 * time.Second

// New builds the playground's http.Handler. public serves the static UI at
// "/"; POST /api/eval evaluates the request body as JavaScript.
func New(public fs.FS) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/eval", func(w http.ResponseWriter, req *http.Request) {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(req.Context(), evalTimeout)
		defer cancel()

		result, err := jsruntime.Eval(ctx, string(b))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(result))
	})

	mux.Handle("/", http.FileServerFS(public))

	return mux
}
