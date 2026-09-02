# wfx UI

Configure the standalone UI by editing `loadConfig` in `index.html`:

```js
function loadConfig() {
  return {
    wfx_url: "https://example.com/api/wfx/v1",
    base_path: "/ui",
    oauth: {
      issuer: "https://identity-provider.example.com",
      client_id: "wfx-ui",
      scope: "openid email profile offline_access",
    },
  };
}
```

Set `oauth` to `null` to disable authentication.
When OAuth is enabled, `issuer` and `client_id` are required. `scope` defaults to `openid email profile`.
Automatic token refresh requires the identity provider's offline-access scope, usually `offline_access`,
so include it in `scope` as shown above.
