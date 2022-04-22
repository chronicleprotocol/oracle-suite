{ pkgs ? import <nixpkgs> { }, oracle-suite ? pkgs.callPackage ./default.nix { } }:
pkgs.mkShell { buildInputs = [ pkgs.jq oracle-suite ]; }
