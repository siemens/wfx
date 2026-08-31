// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import assert from "node:assert/strict";
import { mock } from "node:test";

let manager;
class UserManager {
  constructor(settings) {
    this.settings = settings;
    this.events = {
      addUserLoaded: (callback) => (this.userLoaded = callback),
      addUserUnloaded: (callback) => (this.userUnloaded = callback),
      addAccessTokenExpired: (callback) => (this.accessTokenExpired = callback),
      addSilentRenewError: (callback) => (this.silentRenewError = callback),
    };
    this.getUser = mock.fn();
    this.signinRedirect = mock.fn();
    this.signinRedirectCallback = mock.fn();
    manager = this;
  }
}

mock.module("oidc-client-ts", { exports: { UserManager } });

let replacedWith;
globalThis.window = {
  location: {
    origin: "https://wfx.example",
    pathname: "/ui/",
    search: "",
    href: "https://wfx.example/ui/",
  },
  history: {
    replaceState: (_, __, url) => {
      replacedWith = url;
    },
  },
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
assert.equal(manager, undefined);

const config = {
  base_path: "/ui",
  oauth: new Some({
    issuer: "https://id.example/",
    client_id: "wfx-ui",
    scope: "openid profile",
  }),
};
ready = false;
await auth.authenticate(config, () => {
  ready = true;
});
assert.deepEqual(manager.settings, {
  authority: "https://id.example",
  client_id: "wfx-ui",
  redirect_uri: "https://wfx.example/ui/",
  response_type: "code",
  scope: "openid profile",
  automaticSilentRenew: true,
});
assert.equal(manager.getUser.mock.callCount(), 1);
assert.deepEqual(manager.signinRedirect.mock.calls[0].arguments, [
  { state: "https://wfx.example/ui/" },
]);
assert.equal(ready, false);
assert.equal(auth.accessToken(), "");

const currentUser = { access_token: "stored", expired: false };
manager.getUser.mock.mockImplementationOnce(async () => currentUser);
await auth.authenticate(config, () => {
  ready = true;
});
assert.equal(ready, true);
assert.equal(auth.accessToken(), "stored");

manager.userLoaded({ access_token: "renewed" });
assert.equal(auth.accessToken(), "renewed");
manager.accessTokenExpired();
assert.equal(auth.accessToken(), "");

const callbackUser = {
  access_token: "callback",
  expired: false,
  state: "https://wfx.example/ui/jobs",
};
manager.signinRedirectCallback.mock.mockImplementationOnce(
  async () => callbackUser,
);
window.location.search = "?code=code&state=state";
ready = false;
await auth.authenticate(config, () => {
  ready = true;
});
assert.equal(manager.signinRedirectCallback.mock.callCount(), 1);
assert.equal(manager.getUser.mock.callCount(), 2);
assert.equal(auth.accessToken(), "callback");
assert.equal(replacedWith, callbackUser.state);
assert.equal(ready, true);
