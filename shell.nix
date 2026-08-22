# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
{
  pkgs ? import <nixpkgs> { },
}:

let
  # Evaluate devenv.nix without the devenv framework so shell.nix and devenv share a single package list.
  devenv = import ./devenv.nix {
    inherit pkgs;
    lib = pkgs.lib;
    inputs = { };
    config.devenv.root = toString ./.;
  };

  profilePackages = pkgs.lib.flatten (
    pkgs.lib.mapAttrsToList (
      _: profile: pkgs.lib.mapAttrsToList (_: module: module.packages or [ ]) profile
    ) devenv.profiles
  );
in

with pkgs;

mkShell {
  nativeBuildInputs = devenv.packages ++ profilePackages;
  shellHook = devenv.enterShell;
}
