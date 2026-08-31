// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import assert from "node:assert/strict";

const stored = new Map();
let scheduledTimer;
globalThis.setTimeout = (callback, delay) => {
  scheduledTimer = { callback, delay, unref() {} };
  return scheduledTimer;
};
globalThis.clearTimeout = () => {};
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
let tokenResponses = [];
globalThis.fetch = async (url, options) => {
  requests.push({ url, options });
  if (url === discoveryUrl) {
    return { ok: true, json: async () => provider };
  }
  if (url === provider.token_endpoint) {
    return {
      ok: true,
      json: async () => tokenResponses.shift(),
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
tokenResponses.push({
  access_token: "secret",
  refresh_token: "refresh-secret",
  expires_in: 60,
});
await auth.authenticate(config, () => {
  ready = true;
});
assert.equal(requests[2].url, provider.token_endpoint);
assert.equal(requests[2].options.method, "POST");
assert.equal(requests[2].options.body.get("grant_type"), "authorization_code");
assert.equal(ready, true);
assert.equal(auth.accessToken(), "secret");
const session = JSON.parse(stored.get("wfx.oauth"));
assert.equal(session.refreshToken, "refresh-secret");
assert.equal(session.tokenEndpoint, provider.token_endpoint);
assert.ok(scheduledTimer.delay > 0);

tokenResponses.push({ access_token: "scheduled-refresh", expires_in: 60 });
await scheduledTimer.callback();
assert.equal(auth.accessToken(), "scheduled-refresh");
assert.equal(requests.at(-1).options.body.get("grant_type"), "refresh_token");

const refreshedSession = JSON.parse(stored.get("wfx.oauth"));
refreshedSession.expiresAt = Date.now() - 1;
stored.set("wfx.oauth", JSON.stringify(refreshedSession));
window.location.search = "";
tokenResponses.push({ access_token: "refreshed", expires_in: 60 });
ready = false;
await auth.authenticate(config, () => {
  ready = true;
});
const refreshRequest = requests.at(-1);
assert.equal(refreshRequest.url, provider.token_endpoint);
assert.equal(refreshRequest.options.body.get("grant_type"), "refresh_token");
assert.equal(
  refreshRequest.options.body.get("refresh_token"),
  "refresh-secret",
);
assert.equal(refreshRequest.options.body.get("client_id"), "wfx-ui");
assert.equal(auth.accessToken(), "refreshed");
assert.equal(ready, true);
assert.equal(
  JSON.parse(stored.get("wfx.oauth")).refreshToken,
  "refresh-secret",
);
