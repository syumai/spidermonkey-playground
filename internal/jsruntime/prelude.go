package jsruntime

import _ "embed"

// preludeJS polyfills the subset of Web APIs that hono (and general playground
// code) expects to find on globalThis: Headers, Request, Response,
// URLSearchParams, URL, and no-op addEventListener/removeEventListener/
// dispatchEvent. It is evaluated once per JS instance before user code runs.
// fetch() itself is intentionally not implemented.
//
//go:embed js/prelude.js
var preludeJS string
