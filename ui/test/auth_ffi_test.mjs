// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import assert from "node:assert/strict";

const stored = new Map();
globalThis.sessionStorage = {
  getItem: (key) => stored.get(key) ?? null,
  setItem: (key, value) => stored.set(key, value),
  removeItem: (key) => stored.delete(key),
};
globalThis.btoa = (value) => Buffer.from(value, "binary").toString("base64");
Object.defineProperty(globalThis, "crypto", {
  value: {
    getRandomValues: (bytes) => bytes.fill(1),
    subtle: {
      digest: async (_, bytes) => bytes,
    },
  },
  configurable: true,
});
globalThis.window = {
  location: {
    origin: "https://wfx.example",
    pathname: "/ui",
    search: "",
    href: "https://wfx.example/ui",
    assign: (url) => {
      window.redirectedTo = url;
    },
  },
  history: { replaceState() {} },
};

const discoveryUrl = "https://id.example/.well-known/openid-configuration";
const provider = {
  authorization_endpoint: "https://id.example/authorize",
  token_endpoint: "https://id.example/token",
};
const requests = [];
globalThis.fetch = async (url, options) => {
  requests.push({ url, options });
  if (url === discoveryUrl) {
    return { ok: true, json: async () => provider };
  }
  if (url === provider.token_endpoint) {
    return {
      ok: true,
      json: async () => ({ access_token: "secret", expires_in: 60 }),
    };
  }
  throw new Error(`Unexpected request to ${url}`);
};

const { None, Some } = await import(
  "../build/dev/javascript/gleam_stdlib/gleam/option.mjs"
);
const auth = await import("../src/auth.ffi.mjs");
let ready = false;
await auth.authenticate({ base_path: "/ui", oauth: new None() }, () => {
  ready = true;
});
assert.equal(ready, true);
assert.equal(requests.length, 0);

const config = {
  base_path: "/ui",
  oauth: new Some({
    issuer_url: "https://id.example/",
    client_id: "wfx-ui",
    scope: "openid profile",
  }),
};
ready = false;
await auth.authenticate(config, () => {
  ready = true;
});
assert.equal(requests[0].url, discoveryUrl);
assert.equal(ready, false);
const redirect = new URL(window.redirectedTo);
assert.equal(
  redirect.origin + redirect.pathname,
  provider.authorization_endpoint,
);
assert.equal(redirect.searchParams.get("response_type"), "code");
assert.equal(redirect.searchParams.get("client_id"), "wfx-ui");
assert.equal(
  redirect.searchParams.get("redirect_uri"),
  "https://wfx.example/ui/",
);
assert.equal(redirect.searchParams.get("scope"), "openid profile");
assert.equal(redirect.searchParams.get("code_challenge_method"), "S256");

const state = JSON.parse(stored.get("wfx.oauth")).state;
window.location.search = `?code=code&state=${state}`;
await auth.authenticate(config, () => {
  ready = true;
});
assert.equal(requests[2].url, provider.token_endpoint);
assert.equal(requests[2].options.method, "POST");
assert.equal(ready, true);
assert.equal(auth.accessToken(), "secret");
