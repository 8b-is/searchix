{
  pkgs ? (
    let
      inherit (builtins) fetchTree fromJSON readFile;
      inherit ((fromJSON (readFile ../flake.lock)).nodes) nixpkgs gomod2nix;
    in
    import (fetchTree nixpkgs.locked) {
      overlays = [
        (import "${fetchTree gomod2nix.locked}/overlay.nix")
      ];
    }
  ),
  buildGoApplication ? pkgs.buildGoApplication,
  css,
  lib,
}:
let
  version = "0.4.5";
in
buildGoApplication {
  pname = "searchix";
  inherit version;
  src =
    with pkgs.lib.fileset;
    toSource {
      root = ../.;
      fileset = difference (unions [
        ../go.mod
        ../go.sum
        ../internal
        ../frontend
        ../cmd
        ../web
      ]) (maybeMissing ../frontend/static/base.css);
    };

  subPackages = [ "cmd/searchix-web" ];

  patchPhase = ''
    rm -f frontend/static/base.css
    cp ${css} frontend/static/base.css
  '';
  tags = [ "embed" ];
  ldflags = [
    "-X"
    "alin.ovh/searchix/internal/config.Version=${version}"
  ];

  postInstall = ''
    $out/bin/searchix-web generate-error-page --outdir $out/lib/searchix/
  '';

  modules = ../gomod2nix.toml;

  doCheck = false;

  meta = with lib; {
    description = "Search tool for options and packages in the NixOS ecosystem";
    mainProgram = "searchix-web";
    license = licenses.mit;
    platforms = platforms.all;
  };
}
