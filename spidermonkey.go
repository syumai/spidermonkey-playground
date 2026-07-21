package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"github.com/goccy/go-spidermonkey"
	"github.com/google/uuid"
)

func Eval(code string) (string, error) {
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := js.Eval(ctx, code)
	if err != nil {
		return "", err
	}
	if result.Error != nil {
		return "", result.Error
	}
	return buf.String(), nil
}

// defineCrypto defines the crypto object on the JS instance's global object.
func defineCrypto(js *spidermonkey.JS) error {
	cryptoObj, err := js.NewObject()
	if err != nil {
		return err
	}
	err = cryptoObj.DefineFunc("randomUUID",
		func(cfg spidermonkey.Config, args []spidermonkey.Value) (spidermonkey.Value, error) {
			// UUID v4, same as the Web standard crypto.randomUUID().
			u, err := uuid.NewRandom()
			if err != nil {
				return nil, err
			}
			return spidermonkey.ValueOf(u.String()), nil
		})
	if err != nil {
		return err
	}
	return js.Global().Set("crypto", cryptoObj)
}

// defineConsole defines print and the console object on the JS instance's
// global object. The engine is pure ECMA-262 and has no built-in console;
// output goes to Config.Stdout / Config.Stderr.
func defineConsole(js *spidermonkey.JS) error {
	writer := func(w func(cfg spidermonkey.Config) io.Writer) spidermonkey.Func {
		return func(cfg spidermonkey.Config, args []spidermonkey.Value) (spidermonkey.Value, error) {
			var sb strings.Builder
			for i, a := range args {
				if i > 0 {
					sb.WriteByte(' ')
				}
				sb.WriteString(a.String())
			}
			sb.WriteByte('\n')
			if _, err := io.WriteString(w(cfg), sb.String()); err != nil {
				return nil, err
			}
			return spidermonkey.Undefined(), nil
		}
	}
	stdout := func(cfg spidermonkey.Config) io.Writer { return cfg.Stdout }
	stderr := func(cfg spidermonkey.Config) io.Writer { return cfg.Stderr }

	if err := js.Global().DefineFunc("print", writer(stdout)); err != nil {
		return err
	}
	consoleObj, err := js.NewObject()
	if err != nil {
		return err
	}
	for _, m := range []struct {
		name string
		w    func(cfg spidermonkey.Config) io.Writer
	}{
		{"log", stdout},
		{"info", stdout},
		{"warn", stderr},
		{"error", stderr},
	} {
		if err := consoleObj.DefineFunc(m.name, writer(m.w)); err != nil {
			return err
		}
	}
	return js.Global().Set("console", consoleObj)
}
