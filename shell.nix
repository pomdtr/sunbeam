{ nixpkgs ? import <nixpkgs> {  } }:
 
let
    pkgs = import (builtins.fetchTarball {
        url = "https://github.com/NixOS/nixpkgs/archive/ee01de29d2f58d56b1be4ae24c24bd91c5380cea.tar.gz";
    }) {};
    golang = pkgs.go_1_19;
in
  nixpkgs.stdenv.mkDerivation {
    name = "env";
    buildInputs = [
      golang
      nixpkgs.python3
      nixpkgs.gh
      nixpkgs.nodePackages.zx
      nixpkgs.nodejs-18_x
    ];
  }