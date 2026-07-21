package jsruntime

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestHonoImport is the cheapest possible check that the vendored bundle
// loads and evaluates without hitting a missing global. If this fails, the
// error names the missing identifier (see prelude.js's "risks" note in the
// plan this test came from) and that identifier needs a polyfill.
func TestHonoImport(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	out, err := Eval(ctx, `
import { Hono } from "hono";
console.log(typeof Hono);
`)
	elapsed := time.Since(start)
	t.Logf("hono import took %v", elapsed)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "function" {
		t.Fatalf("unexpected output: %q", out)
	}
}

// TestHonoRouting exercises the router: a static route, a param route, and a
// JSON route, checking status/text/content-type on each response.
func TestHonoRouting(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	out, err := Eval(ctx, `
import { Hono } from "hono";
const app = new Hono();
app.get("/", (c) => c.text("Hello Hono!"));
app.get("/users/:id", (c) => c.json({ id: c.req.param("id") }));

const res1 = await app.fetch(new Request("http://localhost/"));
console.log(res1.status, await res1.text());

const res2 = await app.fetch(new Request("http://localhost/users/42"));
console.log(res2.status, res2.headers.get("content-type"), await res2.text());
`)
	elapsed := time.Since(start)
	t.Logf("hono routing took %v", elapsed)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines of output, got %d: %q", len(lines), out)
	}
	if lines[0] != "200 Hello Hono!" {
		t.Fatalf("unexpected route response: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "200 application/json") || !strings.HasSuffix(lines[1], `{"id":"42"}`) {
		t.Fatalf("unexpected param route response: %q", lines[1])
	}
}

// TestHonoNotFound checks the router's default 404 for an unmatched route.
func TestHonoNotFound(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	out, err := Eval(ctx, `
import { Hono } from "hono";
const app = new Hono();
app.get("/", (c) => c.text("Hello Hono!"));

const res = await app.fetch(new Request("http://localhost/missing"));
console.log(res.status);
`)
	elapsed := time.Since(start)
	t.Logf("hono 404 took %v", elapsed)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "404" {
		t.Fatalf("unexpected output: %q", out)
	}
}
