{ pkgs ? import <nixpkgs> { }
, system ? builtins.currentSystem
, stateVersion ? pkgs.lib.version
, ...
}:
let
  inherit (pkgs) lib;
  nixos = pkgs.nixos ({ lib, pkgs, config, ... }: {
    nixpkgs.hostPlatform = system;
    system.stateVersion = lib.versions.majorMinor stateVersion;
  });

  doc = pkgs.nixosOptionsDoc {
    inherit (nixos) options;
  };
in
doc.optionsJSON
