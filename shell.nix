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
    reuse
    gofumpt

    gnumake
    goreleaser
    zig
    just
    git
    go
  ];

  shellHook = ''
    export GOFLAGS="-tags=sqlite,mysql,postgres,testing"
    export LUA_PATH="$(pwd)/hugo/filters/?.lua;;"
    export PATH="$(pwd):$PATH"
  '';
}
