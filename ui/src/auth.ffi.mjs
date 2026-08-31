// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import { UserManager } from "oidc-client-ts";

let token = "";
let manager;

function callbackUrl(config) {
  return `${window.location.origin}${config.base_path.replace(/\/$/, "")}/`;
}

function setUser(user) {
  token = user?.access_token || "";
}

function getManager(config, oauth) {
  if (manager) return manager;

  manager = new UserManager({
    authority: oauth.issuer_url.replace(/\/$/, ""),
    client_id: oauth.client_id,
    redirect_uri: callbackUrl(config),
    response_type: "code",
    scope: oauth.scope,
    automaticSilentRenew: true,
  });
  manager.events.addUserLoaded(setUser);
  manager.events.addUserUnloaded(() => setUser(null));
  manager.events.addAccessTokenExpired(() => setUser(null));
  manager.events.addSilentRenewError((error) =>
    console.error("OAuth renewal failed:", error),
  );
  return manager;
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

  const userManager = getManager(config, oauth);
  const query = new URLSearchParams(window.location.search);
  const user =
    query.has("code") || query.has("error")
      ? await userManager.signinRedirectCallback()
      : await userManager.getUser();

  if (!user || user.expired) {
    setUser(null);
    await userManager.signinRedirect({ state: window.location.href });
    return;
  }

  setUser(user);
  if (query.has("code") || query.has("error")) {
    window.history.replaceState({}, "", user.state || window.location.pathname);
  }
  onReady();
}

export function accessToken() {
  return token;
}
