// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
let token = "";
let refreshTimer;
let renewal;

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

async function requestToken(tokenEndpoint, parameters) {
  const response = await fetch(tokenEndpoint, {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/x-www-form-urlencoded",
    },
    body: new URLSearchParams(parameters),
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

async function authorize(config, oauth, provider) {
  provider ||= await loadProvider(oauth);
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

function storeToken(config, oauth, result, previous, tokenEndpoint) {
  const expiresIn = result.expires_in;
  if (
    expiresIn != null &&
    (!Number.isFinite(Number(expiresIn)) || Number(expiresIn) <= 0)
  ) {
    throw new Error("OAuth token response has invalid expires_in");
  }
  if (
    result.refresh_token != null &&
    typeof result.refresh_token !== "string"
  ) {
    throw new Error("OAuth token response has invalid refresh_token");
  }

  token = result.access_token;
  const session = {
    token,
    expiresAt: expiresIn == null ? null : Date.now() + Number(expiresIn) * 1000,
    refreshToken: result.refresh_token || previous?.refreshToken || null,
    tokenEndpoint,
  };
  writeSession(session);
  scheduleRenewal(config, oauth, session);
}

function scheduleRenewal(config, oauth, session) {
  clearTimeout(refreshTimer);
  if (!session.expiresAt) return;

  const remaining = Math.max(0, session.expiresAt - Date.now());
  const delay = Math.min(
    remaining > 60_000 ? remaining - 30_000 : remaining * 0.8,
    2_147_483_647,
  );
  refreshTimer = setTimeout(
    () =>
      renew(config, oauth, session).catch((error) => {
        console.error("OAuth renewal failed:", error);
      }),
    delay,
  );
  refreshTimer?.unref?.();
}

function renew(config, oauth, session) {
  if (renewal) return renewal;

  renewal = (async () => {
    try {
      if (!session.refreshToken) {
        token = "";
        await authorize(config, oauth);
        return;
      }
      try {
        const tokenEndpoint =
          session.tokenEndpoint || (await loadProvider(oauth)).token_endpoint;
        const result = await requestToken(tokenEndpoint, {
          grant_type: "refresh_token",
          refresh_token: session.refreshToken,
          client_id: oauth.client_id,
        });
        storeToken(config, oauth, result, session, tokenEndpoint);
      } catch (error) {
        token = "";
        await authorize(config, oauth);
        throw error;
      }
    } finally {
      renewal = null;
    }
  })();
  return renewal;
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
    const result = await requestToken(provider.token_endpoint, {
      grant_type: "authorization_code",
      code,
      client_id: oauth.client_id,
      redirect_uri: callbackUrl(config),
      code_verifier: session.verifier,
    });
    storeToken(config, oauth, result, session, provider.token_endpoint);
    clearCallbackParams(session.returnUrl);
    onReady();
    return;
  }

  if (
    session?.token &&
    (!session.expiresAt || session.expiresAt > Date.now())
  ) {
    token = session.token;
    scheduleRenewal(config, oauth, session);
    onReady();
    return;
  }

  if (session?.refreshToken) {
    await renew(config, oauth, session);
    onReady();
    return;
  }

  await authorize(config, oauth);
}

export function accessToken() {
  return token;
}
