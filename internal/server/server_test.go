package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>playground</html>")},
	}
}

func TestEvalOK(t *testing.T) {
	srv := httptest.NewServer(New(testFS()))
	defer srv.Close()

	res, err := http.Post(srv.URL+"/api/eval", "text/plain", strings.NewReader(`console.log("hi")`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(body)) != "hi" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestEvalError(t *testing.T) {
	srv := httptest.NewServer(New(testFS()))
	defer srv.Close()

	res, err := http.Post(srv.URL+"/api/eval", "text/plain", strings.NewReader(`throw new Error("boom")`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "boom") {
		t.Fatalf("expected body to mention boom, got: %q", body)
	}
}

func TestStaticIndex(t *testing.T) {
	srv := httptest.NewServer(New(testFS()))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(body)) != "<html>playground</html>" {
		t.Fatalf("unexpected body: %q", body)
	}
}
