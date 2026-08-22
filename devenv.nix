# SPDX-FileCopyrightText: 2026 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
{
  pkgs,
  ...
}:

{
  packages = [
    pkgs.git
    pkgs.just
    pkgs.fd
  ];

  env.GOFLAGS = "-tags=testing";

  profiles = {
    fullstack.extends = [
      "backend"
      "frontend"
      "ci"
    ];

    backend.module = {
      packages = [
        pkgs.goreleaser
        pkgs.zig
        pkgs.flatbuffers
        pkgs.gnumake
        pkgs.gnused

        pkgs.go
        pkgs.go-tools
        pkgs.gofumpt
        pkgs.gopls
        pkgs.reftools
        pkgs.golangci-lint
      ];
    };

    frontend.module = {
      env.GOFLAGS = "-tags=testing,ui";
      packages = [
        pkgs.gleam
        pkgs.beamPackages.rebar3
        pkgs.beamPackages.erlang
        pkgs.inotify-tools
        pkgs.nodejs
        pkgs.bun
        pkgs.tailwindcss_4

        pkgs.openssl
        pkgs.zstd
      ];
    };

    pages.module = {
      packages = [
        pkgs.hugo
        pkgs.lychee
        pkgs.python3
      ];
    };

    ci.module = {
      packages = [
        pkgs.reuse
        pkgs.biome
      ];
    };
  };

  enterShell = ''
    set -a; source .env; set +a
    ${pkgs.prek}/bin/prek install --overwrite
  '';
}
