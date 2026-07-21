package jsruntime

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"
)

var uuidV4Re = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestEvalRandomUUID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out1, err := Eval(ctx, `console.log(crypto.randomUUID())`)
	if err != nil {
		t.Fatal(err)
	}
	out2, err := Eval(ctx, `console.log(crypto.randomUUID())`)
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := Eval(ctx, `console.log("Hello")`)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "Hello" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestEvalThrow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Under EvalModule a guest throw surfaces as ModuleResult.Error, but the
	// assertion (error message contains the thrown text) is the same as it
	// was for the classic-script Eval path.
	_, err := Eval(ctx, `throw new Error("boom")`)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected boom error, got: %v", err)
	}
}

func TestEvalTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	_, err := Eval(ctx, `while(true){}`)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if elapsed > 5*time.Second {
		t.Fatalf("interrupt took too long: %v", elapsed)
	}
	t.Logf("interrupted after %v with error: %v", elapsed, err)
}

func TestEvalTopLevelAwait(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := Eval(ctx, `const value = await Promise.resolve(42);
console.log(value);`)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "42" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestEvalStrictMode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Modules are always strict; an undeclared assignment must throw instead
	// of silently creating a global.
	_, err := Eval(ctx, `undeclared = 1;`)
	if err == nil {
		t.Fatal("expected a strict-mode error, got nil")
	}
	t.Logf("got expected error: %v", err)
}

func TestEvalUnknownModule(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := Eval(ctx, `import "lodash";`)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "lodash") || !strings.Contains(err.Error(), `only "hono" is available`) {
		t.Fatalf("error should name the unresolved module and the only available one, got: %v", err)
	}
}
