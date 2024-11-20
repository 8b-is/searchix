{ pkgs ? (
    let
      inherit (builtins) fetchTree fromJSON readFile;
      inherit ((fromJSON (readFile ../flake.lock)).nodes) nixpkgs gomod2nix;
    in
    import (fetchTree nixpkgs.locked) {
      overlays = [
        (import "${fetchTree gomod2nix.locked}/overlay.nix")
      ];
    }
  )
, buildGoApplication ? pkgs.buildGoApplication
, css
}:
let
  version = "0.1.11";
in
buildGoApplication {
  pname = "searchix";
  inherit version;
  src = with pkgs.lib.fileset; toSource {
    root = ../.;
    fileset = difference
      (unions [
        ../go.mod
        ../go.sum
        ../searchix.go
        ../internal
        ../frontend
        ../cmd
      ])
      (maybeMissing ../frontend/static/base.css);
  };

  subPackages = [ "cmd/searchix-web" ];

  patchPhase = ''
    rm -f frontend/static/base.css
    cp ${css} frontend/static/base.css
    ${pkgs.templ}/bin/templ generate
  '';
  tags = [ "embed" ];
  ldflags = [
    "-s"
    "-w"
    "-X"
    "go.alanpearce.eu/searchix/internal/config.Version=${version}"
  ];

  modules = ../gomod2nix.toml;
}
