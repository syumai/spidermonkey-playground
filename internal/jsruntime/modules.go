package jsruntime

import (
	_ "embed"
	"fmt"

	"github.com/goccy/go-spidermonkey"
)

// honoBundle is the vendored jsDelivr ESM build of hono. See js/LICENSE-hono
// for its license and README.md for provenance and how to re-fetch it.
//
//go:embed js/hono.bundle.mjs
var honoBundle string

// moduleLoader resolves the bare specifier "hono" to the vendored bundle. It
// is the only module available to guest code; anything else is a hard
// failure.
func moduleLoader(cfg spidermonkey.Config, specifier, referrer string) (string, error) {
	switch specifier {
	case "hono":
		return honoBundle, nil
	}
	return "", fmt.Errorf("cannot resolve module %q (imported from %q): only \"hono\" is available", specifier, referrer)
}
