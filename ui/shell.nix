# SPDX-FileCopyrightText: 2025 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
let
  pkgs = import <nixpkgs> {
    config = { };
    overlays = [ ];
  };
in
pkgs.mkShell {
  packages = with pkgs; [
    gleam
    beamPackages.rebar3
    inotify-tools
    nodePackages.npm
  ];
}
