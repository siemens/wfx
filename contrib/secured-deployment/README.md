# Secured local deployment

**Warning**: This deployment is not suitable for production. However, it serves as a PoC for how to securely deploy wfx.

## Architecture

```text
       Browser / wfxctl
       (HTTPS :8443)
             |
             v
    +-------------------------------------------------------+
    |                     nginx (:8443)   <-----------------+
    |  - terminates TLS                                     |
    |  - checks token audience per vhost via auth_request   |
    |  - requires admin group + scopes on northbound API    |
    |  - injects X-Client-Id header for southbound jobs     |
    +---+-------------------+-------------------+-----------+
        |                   |                   |           ^
        | :9091             | /oauth2/auth      |           |
        v                   v (:4180)           |           | OIDC discovery
  +------------+     +--------------+           |           | & JWKS via
  |  Authelia  |     | oauth2-proxy |-----------+-----------+ https://authelia.localhost:8443
  |  (OIDC)    |     | (validates   |           |
  +-----+------+     |  JWT/session)|           |
                     +--------------+           |
                                                |
                                                |
                                                |
                                                |
                                                |
                                                |
        +---------------------------------------+
        | route
        |
        |--> wfx-mgmt.localhost:8443 /ui/ ---> UI_DIST directory
        |    (static UI, out-of-tree)
        |
        |--> wfx-mgmt.localhost:8443 --------> wfx northbound (:8081)
        |    (admin jobs/workflows)
        |
        +--> wfx.localhost:8443 -------------> wfx southbound (:8080)
             (client jobs scoped by X-Client-Id)
```

Local test stack:

- nginx terminates TLS on port `8443` and directly serves the static UI from `/ui/`
- Authelia provides OIDC at `https://authelia.localhost:8443`, with Redis-backed sessions
- wfx runs without UI and serves northbound (`:8081`) and southbound (`:8080`) APIs
- `wfx-northbound-api` audience bearer tokens are authorized by oauth2-proxy, then forwarded to northbound port `8081`
- `wfx-southbound-api` audience bearer tokens from the `wfx-devices` client are authorized by oauth2-proxy, then forwarded to southbound port `8080`
- Devices use the OAuth 2.0 device authorization grant (one public `wfx-devices` client) with one Authelia user per device
- Device users carry a `wfx_client_id` attribute; Authelia signs it into the access token as `https://siemens.github.io/wfx/client-id` (RFC 9068 access token)
- nginx rejects tokens whose audience does not match the requested virtual host
- nginx requires the `wfx` group (or the `wfxctl` client credential) and a valid scope for northbound API requests
- nginx requires `read:jobs` or `write:jobs` for GET/HEAD requests, and `write:jobs` for POST/PUT/PATCH/DELETE requests on protected API routes
- Admin client (`wfxctl`) has no device ID injected into `X-Client-Id` on northbound requests
- southbound device IDs come from the signed `https://siemens.github.io/wfx/client-id` claim, never from `client_id` (shared by all devices)
- oauth2-proxy injects the device claim as `X-Wfx-Device-Id`; nginx maps it to `X-Client-Id` before proxying southbound API requests, limiting access to that device's jobs
- unauthenticated API calls get `401`; UI starts OIDC authorization-code flow with PKCE
- supporting services listen only on loopback behind nginx; local wfx listens on all interfaces so nginx container can reach it

## Authorization behavior

Authelia does not restrict token issuance by group, so a device user can still sign in to the
static UI and obtain a `wfx-ui` token. Nginx enforces authorization at the API boundary:

- Northbound user tokens must contain the signed `wfx` group; the `wfxctl` client-credentials
  token is separately allowed for admin CLI use.
- Protected API requests must also carry `read:jobs` or `write:jobs` for GET/HEAD, and
  `write:jobs` for POST/PUT/PATCH/DELETE.
- Device users have no `wfx` group, so their northbound API requests are denied even if their
  token has the correct audience and scopes. The static UI shell may load, but cannot fetch
  management data.
- Per-device job isolation applies to the southbound API; northbound management access is
  admin-only. This PoC therefore does not grant client A northbound access to client B's jobs.

## Start

Start devenv profile (from the git root):

```sh
devenv --profile secured up
```

With the stack running, verify northbound group and scope enforcement from this directory:

```sh
just check-authorization
```

Or start the stack with Docker Compose:

```sh
just bootstrap
just build-ui
docker compose up -d
```


`bootstrap.py` generates local CA, TLS certificate, and OIDC signing certificates.

Launch an isolated Firefox profile which trusts this CA after stack becomes ready:

```sh
just browser
```

Alternatively, trust `ca.pem` in another browser or pass it to clients with `--cacert`.
`*.localhost` names resolve to loopback on common systems; otherwise add `wfx-mgmt.localhost`,
`wfx.localhost`, and `authelia.localhost` to `/etc/hosts`.

## Clients and Users

- `wfx-admin`: web user for wfx-ui
- `wfxctl`: admin CLI client (`client_credentials` grant) for northbound API
- `wfx-devices`: single public OIDC client for all devices (`device_code` + `refresh_token` grants)
- `wfx-client-1` / `wfx-client-2`: Authelia users, one per device; `wfx_client_id` is
  `3eb9312f-1d86-4669-a3e7-ccc69edb1dc0` / `78b062b5-0f10-48d3-bfa5-6a79d44a6d8b`

Device user passwords are generated by `bootstrap.py` into `state/wfx-client-N-password`.
Enrolling a device requires a login as its user (device authorization grant); afterwards
`state/wfx-client-N-tokens.json` holds the refresh token and tokens are obtained without a browser.

## Web UI

Open <https://wfx-mgmt.localhost:8443/ui/>.
Login as `wfx-admin` user (password: `wfx-admin`).

## Endpoints

- Northbound API: `https://wfx-mgmt.localhost:8443/api/wfx/v1/`
- Southbound API: `https://wfx.localhost:8443/api/wfx/v1/`

## Example: client-scoped job visibility

Create a workflow and one job per client on the northbound API with the admin token (which has no
client ID constraint on northbound), then query the southbound API with
each client token. Each client user must see only its own job.

```sh
# 1. Obtain access token for the admin client (pure browserless via client_credentials).
just get-wfx-admin-tokens >wfx-admin-tokens.json
WFX_ADMIN_TOKEN="$(jq -r .access_token wfx-admin-tokens.json)"
alias wfxctl-north="wfxctl --host https://wfx-mgmt.localhost:8443 --tls-ca state/certs/ca.pem --header \"Authorization: Bearer $WFX_ADMIN_TOKEN\""

# 2. Create a workflow.
wfxctl-north workflow create ../../workflow/dau/wfx.workflow.dau.direct.yml

# 3. Create one job for each client.
echo '{ "title": "Job for wfx-client-1" }' | wfxctl-north job create --client-id=3eb9312f-1d86-4669-a3e7-ccc69edb1dc0 --workflow=wfx.workflow.dau.direct -
echo '{ "title": "Job for wfx-client-2" }' | wfxctl-north job create --client-id=78b062b5-0f10-48d3-bfa5-6a79d44a6d8b --workflow=wfx.workflow.dau.direct -

# 4. Obtain each client device's access token. The first call enrolls the
# device (device authorization grant): open the printed URL and log in as
# wfx-client-1 with the password from state/wfx-client-1-password. Subsequent
# calls refresh the stored token without a browser.
just get-wfx-client-1-tokens >wfx-client-1-tokens.json
just get-wfx-client-2-tokens >wfx-client-2-tokens.json
WFX_CLIENT_1_TOKEN="$(jq -r .access_token wfx-client-1-tokens.json)"
WFX_CLIENT_2_TOKEN="$(jq -r .access_token wfx-client-2-tokens.json)"

# 5. Inspect token and observe the signed device claim. client_id is the shared
# `wfx-devices` client, so it must not be used for isolation.
jwt decode $WFX_CLIENT_1_TOKEN
# {
#   ...
#   "client_id": "wfx-devices",
#   "https://siemens.github.io/wfx/client-id": "3eb9312f-1d86-4669-a3e7-ccc69edb1dc0"
# }
```

Query southbound API with each token:

```sh
# use admin token to query all jobs on northbound
wfxctl \
  --host https://wfx-mgmt.localhost:8443 \
  --tls-ca state/certs/ca.pem \
  --header "Authorization: Bearer $WFX_ADMIN_TOKEN" \
  --filter '.content[].clientId' \
  --raw \
  job query

# wfx-client-1 sees only its job.
wfxctl \
  --host https://wfx.localhost:8443 \
  --tls-ca state/certs/ca.pem \
  --header "Authorization: Bearer $WFX_CLIENT_1_TOKEN" \
  --filter '.content[].clientId' \
  --raw \
  job query
# 3eb9312f-1d86-4669-a3e7-ccc69edb1dc0

# wfx-client-2 sees only its job.
wfxctl \
  --host https://wfx.localhost:8443 \
  --tls-ca state/certs/ca.pem \
  --header "Authorization: Bearer $WFX_CLIENT_2_TOKEN" \
  --filter '.content[].clientId' \
  --raw \
  job query
# 78b062b5-0f10-48d3-bfa5-6a79d44a6d8b

# wfxctl admin token is rejected by southbound (audience mismatch: wfx-northbound-api).
wfxctl \
  --host https://wfx.localhost:8443 \
  --tls-ca state/certs/ca.pem \
  --header "Authorization: Bearer $WFX_ADMIN_TOKEN" \
  job query
# HTTP 403
```

**Note**: nginx rejects southbound requests without a device claim in the token. oauth2-proxy
extracts the signed `https://siemens.github.io/wfx/client-id` claim into `X-Wfx-Device-Id`, nginx
maps it to `X-Client-Id`, and wfx uses that header to filter jobs belonging to that device, which
are inaccessible to other devices. oauth2-proxy verifies the signed `client_id` claim
(`audienceClaims: client_id`), and the `aud`-based `X-Accel-Redirect` audience check in nginx
ensures the token's audience matches the requested virtual host.

## Cleanup

Delete `./state` to rotate everything. Delete `state/wfx-client-N-tokens.json` to force a
device re-enrollment.
