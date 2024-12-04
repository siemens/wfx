# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
{ pkgs ? import <nixpkgs> { } }:

with pkgs;

mkShell {
  nativeBuildInputs = [
    (python3.withPackages (ps: with ps; [ pyyaml ]))

    hugo
    htmltest

    nodePackages.sql-formatter
    nodePackages.markdown-link-check
    nodePackages.prettier
    nodePackages.markdownlint-cli
    shfmt

    go-swagger
    golangci-lint
    go-tools
    go-mockery
    reuse
    gofumpt

    gnumake
    goreleaser
    syft
    zig_0_13
    just
    git
    go
    flatbuffers
  ];

  shellHook = ''
    export GOFLAGS="-tags=sqlite,mysql,postgres,testing,integration,plugin"
    export LUA_PATH="$(pwd)/hugo/filters/?.lua;;"
    export PATH="$(pwd):$PATH"

    export PGUSER=wfx \
           PGPASSWORD=secret\
           PGHOST=localhost \
           PGPORT=5432      \
           PGDATABASE=wfx   \
           PGSSLMODE=disable

    export MYSQL_USER=root \
           MYSQL_PASSWORD=root \
           MYSQL_ROOT_PASSWORD=root \
           MYSQL_DATABASE=wfx \
           MYSQL_HOST=localhost
  '';
}
