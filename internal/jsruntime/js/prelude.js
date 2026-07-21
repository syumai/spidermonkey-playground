// Web API polyfill for the playground's guest environment. The engine is
// pure ECMA-262 and has none of Headers / Request / Response /
// URLSearchParams / URL built in, but hono's core (and general playground
// code) expects them on globalThis. fetch() itself is intentionally not
// implemented — only the platform classes it and hono need to construct and
// inspect requests/responses in memory.
(() => {
  "use strict";

  // --- UTF-8 helpers (stand-in for TextEncoder/TextDecoder, which the
  // engine also lacks). Both handle the full Unicode range, including
  // surrogate pairs / lone surrogates (encoded as U+FFFD).

  function utf8Encode(str) {
    const bytes = [];
    for (let i = 0; i < str.length; i++) {
      let code = str.codePointAt(i);
      if (code > 0xffff) i++; // consumed a surrogate pair
      if (code < 0x80) {
        bytes.push(code);
      } else if (code < 0x800) {
        bytes.push(0xc0 | (code >> 6), 0x80 | (code & 0x3f));
      } else if (code < 0x10000) {
        bytes.push(0xe0 | (code >> 12), 0x80 | ((code >> 6) & 0x3f), 0x80 | (code & 0x3f));
      } else {
        bytes.push(
          0xf0 | (code >> 18),
          0x80 | ((code >> 12) & 0x3f),
          0x80 | ((code >> 6) & 0x3f),
          0x80 | (code & 0x3f)
        );
      }
    }
    return new Uint8Array(bytes);
  }

  function utf8Decode(bytes) {
    let result = "";
    let i = 0;
    while (i < bytes.length) {
      const b0 = bytes[i];
      let codePoint, len;
      if (b0 < 0x80) {
        codePoint = b0;
        len = 1;
      } else if ((b0 & 0xe0) === 0xc0) {
        codePoint = b0 & 0x1f;
        len = 2;
      } else if ((b0 & 0xf0) === 0xe0) {
        codePoint = b0 & 0x0f;
        len = 3;
      } else if ((b0 & 0xf8) === 0xf0) {
        codePoint = b0 & 0x07;
        len = 4;
      } else {
        result += "�";
        i += 1;
        continue;
      }
      let valid = true;
      for (let j = 1; j < len; j++) {
        const b = bytes[i + j];
        if (b === undefined || (b & 0xc0) !== 0x80) {
          valid = false;
          break;
        }
        codePoint = (codePoint << 6) | (b & 0x3f);
      }
      if (!valid) {
        result += "�";
        i += 1;
        continue;
      }
      result += String.fromCodePoint(codePoint);
      i += len;
    }
    return result;
  }

  // normalizeBody accepts the shapes fetch()-adjacent APIs commonly pass as a
  // body (null, string, Uint8Array, ArrayBuffer) and returns a Uint8Array (or
  // null for no body). Anything else is coerced through String() so
  // constructing a Request/Response never throws on an unusual body.
  function normalizeBody(input) {
    if (input === null || input === undefined) return null;
    if (typeof input === "string") return utf8Encode(input);
    if (input instanceof Uint8Array) return input;
    if (input instanceof ArrayBuffer) return new Uint8Array(input);
    return utf8Encode(String(input));
  }

  // --- Headers ---

  class Headers {
    constructor(init) {
      this._map = new Map(); // lowercase name -> value
      if (init instanceof Headers) {
        for (const [k, v] of init._map) this._map.set(k, v);
      } else if (Array.isArray(init)) {
        for (const pair of init) this.append(pair[0], pair[1]);
      } else if (init && typeof init === "object") {
        for (const k of Object.keys(init)) this.append(k, init[k]);
      }
    }
    get(name) {
      const v = this._map.get(String(name).toLowerCase());
      return v === undefined ? null : v;
    }
    set(name, value) {
      this._map.set(String(name).toLowerCase(), String(value));
    }
    append(name, value) {
      const key = String(name).toLowerCase();
      const existing = this._map.get(key);
      this._map.set(key, existing === undefined ? String(value) : existing + ", " + String(value));
    }
    has(name) {
      return this._map.has(String(name).toLowerCase());
    }
    delete(name) {
      this._map.delete(String(name).toLowerCase());
    }
    forEach(callback, thisArg) {
      for (const [k, v] of this._map) callback.call(thisArg, v, k, this);
    }
    *entries() {
      yield* this._map.entries();
    }
    *keys() {
      yield* this._map.keys();
    }
    *values() {
      yield* this._map.values();
    }
    [Symbol.iterator]() {
      return this.entries();
    }
  }

  // --- Request ---

  class Request {
    constructor(input, init) {
      init = init || {};
      const fromRequest = input instanceof Request;
      this.url = fromRequest ? input.url : String(input);
      this.method = init.method ? String(init.method).toUpperCase() : fromRequest ? input.method : "GET";
      this.headers = new Headers(init.headers !== undefined ? init.headers : fromRequest ? input.headers : undefined);
      const body = init.body !== undefined ? init.body : fromRequest ? input._bodyBytes : null;
      this._bodyBytes = normalizeBody(body);
      this.body = this._bodyBytes;
      this.bodyUsed = false;
    }
    async text() {
      return this._bodyBytes ? utf8Decode(this._bodyBytes) : "";
    }
    async json() {
      return JSON.parse(await this.text());
    }
    async arrayBuffer() {
      return (this._bodyBytes || new Uint8Array(0)).buffer;
    }
    clone() {
      return new Request(this);
    }
  }

  // --- Response ---

  class Response {
    constructor(body, init) {
      init = init || {};
      this.status = init.status !== undefined ? init.status : 200;
      this.statusText = init.statusText !== undefined ? init.statusText : "";
      this.headers = new Headers(init.headers);
      this._bodyBytes = normalizeBody(body === undefined ? null : body);
      this.body = this._bodyBytes;
      this.ok = this.status >= 200 && this.status < 300;
      this.redirected = false;
      this.url = "";
      this.bodyUsed = false;
    }
    async text() {
      return this._bodyBytes ? utf8Decode(this._bodyBytes) : "";
    }
    async json() {
      return JSON.parse(await this.text());
    }
    async arrayBuffer() {
      return (this._bodyBytes || new Uint8Array(0)).buffer;
    }
    clone() {
      const r = new Response(this._bodyBytes, {
        status: this.status,
        statusText: this.statusText,
        headers: this.headers,
      });
      r.url = this.url;
      r.redirected = this.redirected;
      return r;
    }
    static json(data, init) {
      init = init || {};
      const headers = new Headers(init.headers);
      if (!headers.has("content-type")) headers.set("content-type", "application/json");
      return new Response(JSON.stringify(data), {
        status: init.status,
        statusText: init.statusText,
        headers,
      });
    }
    static redirect(url, status) {
      const headers = new Headers();
      headers.set("location", String(url));
      return new Response(null, { status: status === undefined ? 302 : status, headers });
    }
  }

  // --- URLSearchParams ---

  function encodeSearchComponent(s) {
    return encodeURIComponent(s).replace(/%20/g, "+");
  }

  class URLSearchParams {
    constructor(init) {
      this._list = []; // [name, value] pairs, insertion order, duplicates kept
      if (init === undefined || init === null) {
        // empty
      } else if (init instanceof URLSearchParams) {
        for (const pair of init._list) this._list.push([pair[0], pair[1]]);
      } else if (typeof init === "string") {
        let s = init;
        if (s.startsWith("?")) s = s.slice(1);
        if (s.length > 0) {
          for (const part of s.split("&")) {
            if (part === "") continue;
            const eq = part.indexOf("=");
            const rawK = eq === -1 ? part : part.slice(0, eq);
            const rawV = eq === -1 ? "" : part.slice(eq + 1);
            this._list.push([
              decodeURIComponent(rawK.replace(/\+/g, " ")),
              decodeURIComponent(rawV.replace(/\+/g, " ")),
            ]);
          }
        }
      } else if (Array.isArray(init)) {
        for (const pair of init) this._list.push([String(pair[0]), String(pair[1])]);
      } else if (typeof init === "object") {
        for (const k of Object.keys(init)) this._list.push([k, String(init[k])]);
      }
    }
    append(name, value) {
      this._list.push([String(name), String(value)]);
    }
    delete(name) {
      name = String(name);
      this._list = this._list.filter((pair) => pair[0] !== name);
    }
    get(name) {
      name = String(name);
      for (const pair of this._list) if (pair[0] === name) return pair[1];
      return null;
    }
    getAll(name) {
      name = String(name);
      return this._list.filter((pair) => pair[0] === name).map((pair) => pair[1]);
    }
    has(name) {
      name = String(name);
      return this._list.some((pair) => pair[0] === name);
    }
    set(name, value) {
      name = String(name);
      value = String(value);
      let found = false;
      const next = [];
      for (const pair of this._list) {
        if (pair[0] !== name) {
          next.push(pair);
        } else if (!found) {
          next.push([name, value]);
          found = true;
        }
      }
      if (!found) next.push([name, value]);
      this._list = next;
    }
    forEach(callback, thisArg) {
      for (const pair of this._list) callback.call(thisArg, pair[1], pair[0], this);
    }
    *entries() {
      for (const pair of this._list) yield [pair[0], pair[1]];
    }
    *keys() {
      for (const pair of this._list) yield pair[0];
    }
    *values() {
      for (const pair of this._list) yield pair[1];
    }
    [Symbol.iterator]() {
      return this.entries();
    }
    get size() {
      return this._list.length;
    }
    toString() {
      return this._list.map((pair) => encodeSearchComponent(pair[0]) + "=" + encodeSearchComponent(pair[1])).join("&");
    }
  }

  // --- URL ---
  // http(s)-only. base resolution is a simple origin/path join, not a full
  // WHATWG URL-parsing algorithm.

  const URL_RE = /^(https?:)\/\/([^/?#]*)([^?#]*)(\?[^#]*)?(#.*)?$/;

  class URL {
    constructor(input, base) {
      let str = String(input);
      let m = URL_RE.exec(str);
      if (!m && base !== undefined && base !== null) {
        const baseURL = base instanceof URL ? base : new URL(base);
        if (str.startsWith("//")) {
          str = baseURL.protocol + str;
        } else if (str.startsWith("/")) {
          str = baseURL.origin + str;
        } else if (!/^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(str)) {
          const baseDir = baseURL.pathname.replace(/[^/]*$/, "");
          str = baseURL.origin + baseDir + str;
        }
        m = URL_RE.exec(str);
      }
      if (!m) throw new TypeError("Invalid URL: " + String(input));

      this.protocol = m[1];
      const authority = m[2] || "";
      const at = authority.indexOf("@");
      const hostport = at === -1 ? authority : authority.slice(at + 1);
      const colon = hostport.lastIndexOf(":");
      if (colon === -1) {
        this.hostname = hostport;
        this.port = "";
      } else {
        this.hostname = hostport.slice(0, colon);
        this.port = hostport.slice(colon + 1);
      }
      this.host = this.port ? this.hostname + ":" + this.port : this.hostname;
      this.pathname = m[3] && m[3].length > 0 ? m[3] : "/";
      this.search = m[4] || "";
      this.hash = m[5] || "";
      this.origin = this.protocol + "//" + this.host;
      this._searchParams = null;
    }
    get href() {
      return this.origin + this.pathname + this.search + this.hash;
    }
    get searchParams() {
      if (!this._searchParams) this._searchParams = new URLSearchParams(this.search);
      return this._searchParams;
    }
    toString() {
      return this.href;
    }
    toJSON() {
      return this.href;
    }
  }

  Object.assign(globalThis, { Headers, Request, Response, URLSearchParams, URL });

  // Some bundles reference these defensively even when nothing ever
  // dispatches an event in this environment; no-op them for safety.
  globalThis.addEventListener = () => {};
  globalThis.removeEventListener = () => {};
  globalThis.dispatchEvent = () => false;
})();
