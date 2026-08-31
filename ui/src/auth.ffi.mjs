// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
let token = "";

const storageKey = "wfx.oauth";

function callbackUrl(config) {
  return `${window.location.origin}${config.base_path.replace(/\/$/, "")}/`;
}

function encode(bytes) {
  return btoa(String.fromCharCode(...bytes))
    .replaceAll("+", "-")
    .replaceAll("/", "_")
    .replaceAll("=", "");
}

function randomString() {
  const bytes = new Uint8Array(32);
  crypto.getRandomValues(bytes);
  return encode(bytes);
}

async function codeChallenge(verifier) {
  const bytes = new TextEncoder().encode(verifier);
  return encode(new Uint8Array(await crypto.subtle.digest("SHA-256", bytes)));
}

function readSession() {
  try {
    return JSON.parse(sessionStorage.getItem(storageKey));
  } catch {
    return null;
  }
}

function writeSession(value) {
  try {
    sessionStorage.setItem(storageKey, JSON.stringify(value));
  } catch {
    throw new Error("OAuth requires sessionStorage");
  }
}

function clearCallbackParams(returnUrl) {
  window.history.replaceState({}, "", returnUrl || window.location.pathname);
}

async function loadProvider(oauth) {
  const response = await fetch(
    `${oauth.issuer_url.replace(/\/$/, "")}/.well-known/openid-configuration`,
    { headers: { Accept: "application/json" } },
  );
  if (!response.ok) {
    throw new Error(
      `OpenID configuration request failed with status ${response.status}`,
    );
  }
  const provider = await response.json();
  if (!provider.authorization_endpoint || !provider.token_endpoint) {
    throw new Error(
      "OpenID configuration has no authorization_endpoint or token_endpoint",
    );
  }
  return provider;
}

async function requestToken(config, oauth, tokenEndpoint, code, verifier) {
  const body = new URLSearchParams({
    grant_type: "authorization_code",
    code,
    client_id: oauth.client_id,
    redirect_uri: callbackUrl(config),
    code_verifier: verifier,
  });
  const response = await fetch(tokenEndpoint, {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/x-www-form-urlencoded",
    },
    body,
  });
  if (!response.ok) {
    throw new Error(
      `OAuth token request failed with status ${response.status}`,
    );
  }
  const result = await response.json();
  if (!result.access_token)
    throw new Error("OAuth token response has no access_token");
  return result;
}

export async function authenticate(config, onReady) {
  if (!config.oauth[0]) {
    onReady();
    return;
  }
  const oauth = config.oauth[0];
  if (!oauth.issuer_url || !oauth.client_id) {
    throw new Error("OAuth requires issuer_url and client_id");
  }

  const query = new URLSearchParams(window.location.search);
  const session = readSession();
  if (query.has("error")) {
    throw new Error(
      `OAuth authorization failed: ${query.get("error_description") || query.get("error")}`,
    );
  }

  const code = query.get("code");
  if (code) {
    if (!session?.state || query.get("state") !== session.state) {
      clearCallbackParams(session?.returnUrl);
      throw new Error("OAuth state mismatch");
    }
    const provider = await loadProvider(oauth);
    const result = await requestToken(
      config,
      oauth,
      provider.token_endpoint,
      code,
      session.verifier,
    );
    token = result.access_token;
    writeSession({
      token,
      expiresAt: result.expires_in
        ? Date.now() + Number(result.expires_in) * 1000
        : null,
    });
    clearCallbackParams(session.returnUrl);
    onReady();
    return;
  }

  if (
    session?.token &&
    (!session.expiresAt || session.expiresAt > Date.now())
  ) {
    token = session.token;
    onReady();
    return;
  }

  const provider = await loadProvider(oauth);
  const state = randomString();
  const verifier = randomString();
  writeSession({ state, verifier, returnUrl: window.location.href });
  const url = new URL(provider.authorization_endpoint);
  url.searchParams.set("response_type", "code");
  url.searchParams.set("client_id", oauth.client_id);
  url.searchParams.set("redirect_uri", callbackUrl(config));
  url.searchParams.set("code_challenge", await codeChallenge(verifier));
  url.searchParams.set("code_challenge_method", "S256");
  url.searchParams.set("state", state);
  if (oauth.scope) url.searchParams.set("scope", oauth.scope);
  window.location.assign(url);
}

export function accessToken() {
  return token;
}
