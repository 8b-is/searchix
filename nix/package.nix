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
buildGoApplication {
  pname = "searchix";
  version = "0.1";
  src = with pkgs.lib.fileset; toSource {
    root = ../.;
    fileset = difference
      (unions [
        ../go.mod
        ../go.sum
        ../serve
        ../import
        ../internal
        ../frontend
      ])
      (maybeMissing ../frontend/static/base.css);
  };

  patchPhase = ''
    cp ${css} frontend/static/base.css
  '';
  tags = [ "embed" ];
  modules = ../gomod2nix.toml;
}
