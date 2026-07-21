package jsruntime

import (
	"context"
	"strings"
	"testing"
	"time"
)

// evalPolyfill is a small helper: it wraps expr's completion in
// console.log(JSON.stringify(...)) so the assertion just has to compare
// trimmed stdout, without every test needing its own console.log call.
func evalPolyfill(t *testing.T, code string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := Eval(ctx, code)
	if err != nil {
		t.Fatalf("Eval(%q) failed: %v", code, err)
	}
	return strings.TrimSpace(out)
}

func TestHeadersCaseInsensitive(t *testing.T) {
	got := evalPolyfill(t, `
const h = new Headers();
h.set("Content-Type", "text/plain");
console.log(h.get("content-type"), h.has("CONTENT-TYPE"));
`)
	if got != "text/plain true" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestHeadersInitForms(t *testing.T) {
	got := evalPolyfill(t, `
const fromArray = new Headers([["X-A", "1"], ["X-B", "2"]]);
const fromObject = new Headers({ "X-A": "1", "X-B": "2" });
const fromHeaders = new Headers(fromArray);
console.log(fromArray.get("x-a"), fromObject.get("x-b"), fromHeaders.get("x-a"));
`)
	if got != "1 2 1" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestHeadersAppendJoinsWithComma(t *testing.T) {
	got := evalPolyfill(t, `
const h = new Headers();
h.append("X-Multi", "a");
h.append("X-Multi", "b");
console.log(h.get("x-multi"));
`)
	if got != "a, b" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestRequestMethodUppercasedAndDefaultsToGet(t *testing.T) {
	got := evalPolyfill(t, `
const a = new Request("http://example.com/");
const b = new Request("http://example.com/", { method: "post" });
console.log(a.method, b.method);
`)
	if got != "GET POST" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestRequestTextAndJSON(t *testing.T) {
	got := evalPolyfill(t, `
const req = new Request("http://example.com/", { method: "POST", body: JSON.stringify({ a: 1 }) });
const text = await req.text();
const req2 = new Request("http://example.com/", { method: "POST", body: JSON.stringify({ a: 2 }) });
const json = await req2.json();
console.log(text, json.a);
`)
	if got != `{"a":1} 2` {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestResponseDefaultStatus(t *testing.T) {
	got := evalPolyfill(t, `
const res = new Response("hi");
console.log(res.status, res.ok, await res.text());
`)
	if got != "200 true hi" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestResponseJSON(t *testing.T) {
	got := evalPolyfill(t, `
const res = Response.json({ ok: true }, { status: 201 });
console.log(res.status, res.headers.get("content-type"), await res.text());
`)
	if got != `201 application/json {"ok":true}` {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestResponseNotOkForErrorStatus(t *testing.T) {
	got := evalPolyfill(t, `
const res = new Response("nope", { status: 404 });
console.log(res.ok, res.status);
`)
	if got != "false 404" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestURLProperties(t *testing.T) {
	got := evalPolyfill(t, `
const u = new URL("https://user@example.com:8080/path/to/thing?a=1&b=2#frag");
console.log(u.protocol, u.hostname, u.port, u.pathname, u.search, u.hash, u.origin);
`)
	if got != "https: example.com 8080 /path/to/thing ?a=1&b=2 #frag https://example.com:8080" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestURLSearchParams(t *testing.T) {
	got := evalPolyfill(t, `
const u = new URL("https://example.com/?a=1&a=2&b=3");
console.log(u.searchParams.get("a"), u.searchParams.getAll("a").join(","), u.searchParams.get("b"));
`)
	if got != "1 1,2 3" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestURLDefaultPathname(t *testing.T) {
	got := evalPolyfill(t, `
const u = new URL("https://example.com");
console.log(JSON.stringify(u.pathname));
`)
	if got != `"/"` {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestURLInvalidThrowsTypeError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := Eval(ctx, `new URL("not a url");`)
	if err == nil || !strings.Contains(err.Error(), "Invalid URL") {
		t.Fatalf("expected an Invalid URL error, got: %v", err)
	}
}

func TestAddEventListenerIsNoop(t *testing.T) {
	got := evalPolyfill(t, `
addEventListener("fetch", () => {});
removeEventListener("fetch", () => {});
console.log(typeof addEventListener, dispatchEvent("x"));
`)
	if got != "function false" {
		t.Fatalf("unexpected output: %q", got)
	}
}
