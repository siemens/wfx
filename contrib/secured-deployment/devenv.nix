# SPDX-FileCopyrightText: 2026 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
{
  pkgs,
  config,
  ...
}:

let
  root = config.devenv.root;
  state = "${root}/contrib/secured-deployment/state";
in
{
  packages = [
    pkgs.authelia
    pkgs.curl
    pkgs.firefox
    pkgs.jq
    pkgs.nginx
    pkgs.nssTools
    pkgs.oauth2-proxy
    pkgs.oauth2c
  ];

  languages.python = {
    enable = true;
    package = pkgs.python3.withPackages (ps: [
      ps.cryptography
    ]);
  };

  tasks."secured:bootstrap".exec = ''
    set -eu
    cd "${root}/contrib/secured-deployment"
    ${config.languages.python.package}/bin/python3 bootstrap.py --local --state "${state}"
  '';

  tasks."secured:build-ui".exec = ''
    cd "${root}/contrib/secured-deployment"
    ${pkgs.just}/bin/just build-ui
  '';

  processes.wfx = {
    exec = ''
      mkdir -p "${state}"
      cd "${root}"
      exec ${pkgs.go}/bin/go run -tags '!ui' ./cmd/wfx \
        --client-host=http://127.0.0.1:8080 \
        --mgmt-host=http://127.0.0.1:8081 \
        --storage=sqlite \
        --storage-opt="file:${state}/wfx.db?_fk=1&_journal=WAL"
    '';
    ready.http.get = {
      port = 8081;
      path = "/health";
    };
    after = [
      "devenv:processes:oauth2-proxy"
    ];
  };

  processes.authelia = {
    exec = ''
      set -a
      . "${state}/secrets.env"
      set +a
      exec ${pkgs.authelia}/bin/authelia \
        --config.experimental.filters=template \
        --config="${state}/authelia.yml"
    '';
    after = [
      "secured:bootstrap"
      "devenv:processes:redis"
    ];
    ready.http.get = {
      port = 9091;
      path = "/api/health";
    };
  };

  services.redis = {
    enable = true;
    bind = "127.0.0.1";
    port = 6379;
  };

  services.nginx = {
    enable = true;
    httpConfig = builtins.replaceStrings [ "@STATE@" "@UI_DIST@" ] [ state "${root}/ui/dist" ] (
      builtins.readFile ./nginx.conf
    );
  };
  processes.nginx = {
    after = [
      "secured:bootstrap"
      "secured:build-ui"
      "devenv:processes:authelia"
    ];
    ready.exec = ''
      ${pkgs.curl}/bin/curl --fail --silent \
        --cacert ${state}/certs/ca.pem \
        https://authelia.localhost:8443/api/health >/dev/null
    '';
  };

  processes.oauth2-proxy = {
    exec = ''
      set -a
      . "${state}/secrets.env"
      set +a
      # Make nss-myhostname available to OAuth2 Proxy's glibc resolver.
      export LD_LIBRARY_PATH="${pkgs.systemd}/lib''${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
      exec ${pkgs.oauth2-proxy}/bin/oauth2-proxy \
        --config="${state}/oauth2-proxy.cfg" \
        --alpha-config="${state}/oauth2-proxy-alpha.yml"
    '';
    after = [
      "secured:bootstrap"
      "devenv:processes:nginx"
    ];
    ready.http.get = {
      port = 4180;
      path = "/ping";
    };
  };
}
