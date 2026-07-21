package jsruntime

import (
	"io"
	"strings"

	"github.com/goccy/go-spidermonkey"
)

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
