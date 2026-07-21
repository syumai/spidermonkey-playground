// Package jsruntime evaluates JavaScript on the SpiderMonkey engine
// (via github.com/goccy/go-spidermonkey), with console/crypto host bindings,
// a Web API polyfill (see js/prelude.js), and "hono" available to import.
package jsruntime

import (
	"bytes"
	"context"
	"fmt"

	"github.com/goccy/go-spidermonkey"
)

// playgroundSpecifier is the module specifier user code is evaluated under.
// It has no meaning beyond identifying the top-level module to the engine;
// it is never resolved against the filesystem.
const playgroundSpecifier = "playground.js"

// Eval runs code as an ES module (so import/export and top-level await are
// always available) and returns everything written to console/print during
// execution. ctx bounds the whole run, including the prelude and module
// evaluation; a runaway script is interrupted when ctx is done.
func Eval(ctx context.Context, code string) (string, error) {
	var buf bytes.Buffer
	js, err := spidermonkey.New(spidermonkey.Config{Stdout: &buf, Stderr: &buf})
	if err != nil {
		return "", err
	}
	defer js.Close()

	if err := defineCrypto(js); err != nil {
		return "", err
	}
	if err := defineConsole(js); err != nil {
		return "", err
	}
	js.SetModuleLoader(moduleLoader)

	// The prelude is host-provided setup, not user code: a failure here is
	// our bug, not the guest script's, so it gets a distinguishing error
	// wrapper instead of being reported the same way as a user-code failure.
	if preludeResult, err := js.Eval(ctx, preludeJS); err != nil {
		return "", fmt.Errorf("prelude: %w", err)
	} else if preludeResult.Error != nil {
		return "", fmt.Errorf("prelude: %w", preludeResult.Error)
	}

	modResult, err := js.EvalModule(ctx, playgroundSpecifier, code)
	if err != nil {
		return "", err
	}
	if modResult.Error != nil {
		return "", modResult.Error
	}
	return buf.String(), nil
}
