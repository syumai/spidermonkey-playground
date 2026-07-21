# spidermonkey-playground

An HTTP API that evaluates JavaScript using the SpiderMonkey engine (via [goccy/go-spidermonkey](https://github.com/goccy/go-spidermonkey)).

## API

### `POST /api/eval`

Send JavaScript source code as the raw, plain-text request body. The code is run as an ES module (so top-level `import`/`export` and top-level `await` are always available) and its stdout is returned as the plain-text response body.

- On success: `200 OK` with the script's stdout.
- On failure: `500 Internal Server Error` with the error message as plain text.
- Execution is interrupted after approximately 3 seconds.
- `console.log`/`info`/`warn`/`error`, `print`, and `crypto.randomUUID()` are available, as are `Headers`, `Request`, `Response`, `URL`, and `URLSearchParams` (a Web API polyfill — see below). `fetch()` itself is not implemented.
- `import { Hono } from "hono"` is available (a vendored build of [hono](https://hono.dev) — see below). No other module specifier resolves.

Example:

```
curl -X POST http://localhost:3000/api/eval --data 'console.log(1 + 2);'
```

Hono example:

```
curl -X POST http://localhost:3000/api/eval --data 'import { Hono } from "hono";
const app = new Hono();
app.get("/", (c) => c.text("Hi"));
console.log(await (await app.fetch(new Request("http://localhost/"))).text());'
```

## Web API polyfill

The engine (`goccy/go-spidermonkey`) is exactly ECMA-262: it has no built-in `console`, `crypto`, or platform/Web APIs. This project adds:

- `console`/`print`, `crypto.randomUUID()` — defined directly on the Go side (`internal/jsruntime/console.go`, `internal/jsruntime/crypto.go`).
- `Headers`, `Request`, `Response`, `URLSearchParams`, `URL`, and no-op `addEventListener`/`removeEventListener`/`dispatchEvent` — a JavaScript polyfill (`internal/jsruntime/js/prelude.js`), evaluated before every script. This is enough for `hono`'s core router; it is not a complete or spec-exact implementation (e.g. `URL` only parses `http(s)://`, and there is no streaming `body`). `fetch()` itself is intentionally not implemented.

## Vendored hono bundle

`internal/jsruntime/js/hono.bundle.mjs` is hono's official jsDelivr ESM build, fetched once at development time and embedded into the server binary — there is no CDN dependency at runtime.

- Source: `https://cdn.jsdelivr.net/npm/hono@4.12.31/+esm`
- Version: `4.12.31`
- License: MIT — see `internal/jsruntime/js/LICENSE-hono` (fetched from `https://raw.githubusercontent.com/honojs/hono/main/LICENSE`)

To re-fetch (e.g. to bump the version):

```
curl -fsSL -o internal/jsruntime/js/hono.bundle.mjs "https://cdn.jsdelivr.net/npm/hono@<version>/+esm"
curl -fsSL -o internal/jsruntime/js/LICENSE-hono "https://raw.githubusercontent.com/honojs/hono/main/LICENSE"
```

After re-fetching, verify the bundle has no external imports and still exports `Hono`:

```
grep -cE '^import |from"https?:' internal/jsruntime/js/hono.bundle.mjs   # expect 0
tail -c 300 internal/jsruntime/js/hono.bundle.mjs                        # expect an `export { ..., Hono }` at the end
```

## Web Playground

[`public/index.html`](public/index.html) is embedded into the server binary (via `go:embed`) and served at `/`. It provides a simple form for trying the API from the browser: write some JavaScript, run it, and see the output.

## Development

Run the server locally with:

```
go run .
```

This serves both the API and the playground, listening on `$PORT` (default `3000`).

## License

[MIT](LICENSE)
