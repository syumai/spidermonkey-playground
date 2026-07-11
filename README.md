# spidermonkey-playground

An HTTP API that evaluates JavaScript using the SpiderMonkey engine (via [goccy/go-spidermonkey](https://github.com/goccy/go-spidermonkey)).

## API

### `POST /api/eval`

Send JavaScript source code as the raw, plain-text request body. The script is evaluated and its stdout is returned as the plain-text response body.

- On success: `200 OK` with the script's stdout.
- On failure: `500 Internal Server Error` with the error message as plain text.
- Execution is interrupted after approximately 1 second.

Example:

```
curl -X POST http://localhost:3000/api/eval --data 'console.log(1 + 2);'
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
