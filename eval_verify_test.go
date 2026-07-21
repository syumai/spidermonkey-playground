package main

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

var uuidV4Re = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestEvalRandomUUID(t *testing.T) {
	out1, err := Eval(`console.log(crypto.randomUUID())`)
	if err != nil {
		t.Fatal(err)
	}
	out2, err := Eval(`console.log(crypto.randomUUID())`)
	if err != nil {
		t.Fatal(err)
	}
	u1 := strings.TrimSpace(out1)
	u2 := strings.TrimSpace(out2)
	if !uuidV4Re.MatchString(u1) {
		t.Fatalf("not a UUID v4: %q", u1)
	}
	if u1 == u2 {
		t.Fatalf("two calls returned the same UUID: %q", u1)
	}
}

func TestEvalConsoleLog(t *testing.T) {
	out, err := Eval(`console.log("Hello")`)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "Hello" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestEvalThrow(t *testing.T) {
	_, err := Eval(`throw new Error("boom")`)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected boom error, got: %v", err)
	}
}

func TestEvalTimeout(t *testing.T) {
	start := time.Now()
	_, err := Eval(`while(true){}`)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if elapsed > 5*time.Second {
		t.Fatalf("interrupt took too long: %v", elapsed)
	}
	t.Logf("interrupted after %v with error: %v", elapsed, err)
}
