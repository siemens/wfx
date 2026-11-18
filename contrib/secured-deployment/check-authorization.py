#!/usr/bin/env python3
# SPDX-FileCopyrightText: 2026 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0

import base64
import json
from pathlib import Path
import socket
import ssl
import time
import urllib.error
import urllib.request

from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import padding


ROOT = Path(__file__).parent
API_URL = "https://wfx-mgmt.localhost:8443/api/wfx/v1/jobs"

key = serialization.load_pem_private_key(
    (ROOT / "state/certs/oidc-key.pem").read_bytes(), password=None
)
tls = ssl.create_default_context(cafile=str(ROOT / "state/certs/ca.pem"))
getaddrinfo = socket.getaddrinfo
socket.getaddrinfo = lambda host, *args, **kwargs: getaddrinfo(
    "127.0.0.1" if host.endswith(".localhost") else host, *args, **kwargs
)


def b64(value):
    return base64.urlsafe_b64encode(value).rstrip(b"=").decode()


def token(client_id, groups, scopes):
    now = int(time.time())
    header = {"alg": "RS256", "kid": "wfx", "typ": "JWT"}
    claims = {
        "iss": "https://authelia.localhost:8443",
        "sub": "wfx-admin" if groups else "wfx-client-1",
        "email": "wfx-admin@local" if groups else "wfx-client-1@local",
        "client_id": client_id,
        "aud": ["wfx-northbound-api"],
        "scp": scopes,
        "iat": now,
        "exp": now + 60,
    }
    if groups:
        claims["groups"] = groups
    payload = ".".join(
        b64(json.dumps(value, separators=(",", ":")).encode())
        for value in (header, claims)
    )
    signature = key.sign(payload.encode(), padding.PKCS1v15(), hashes.SHA256())
    return f"{payload}.{b64(signature)}"


def expect(label, bearer, expected, method="GET"):
    request = urllib.request.Request(
        API_URL,
        headers={"Authorization": f"Bearer {bearer}"},
        method=method,
    )
    try:
        with urllib.request.urlopen(request, context=tls, timeout=10) as response:
            actual = response.status
    except urllib.error.HTTPError as error:
        actual = error.code
    if actual == expected:
        print(f"✅ OK: {label} (HTTP {actual})")
        return
    print(f"❌ FAILED: {label} (expected HTTP {expected}, got HTTP {actual})")
    raise SystemExit(1)


expect(
    "device user with northbound audience and scopes",
    token("wfx-ui", [], ["read:jobs", "write:jobs"]),
    403,
)
expect("admin group without scope", token("wfx-ui", ["wfx-admin"], []), 403)
expect("admin group with read scope", token("wfx-ui", ["wfx-admin"], ["read:jobs"]), 200)
expect(
    "wfxctl without admin group",
    token("wfxctl", [], ["read:jobs", "write:jobs"]),
    403,
)
expect(
    "wfxctl user with admin group",
    token("wfxctl", ["wfx-admin"], ["read:jobs", "write:jobs"]),
    200,
)
expect(
    "read-only admin cannot POST",
    token("wfx-ui", ["wfx-admin"], ["read:jobs"]),
    403,
    method="POST",
)
expect(
    "write-scoped admin reaches POST handler",
    token("wfx-ui", ["wfx-admin"], ["write:jobs"]),
    400,
    method="POST",
)
