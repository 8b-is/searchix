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
, self
}:
let
  inherit (builtins) concatStringsSep match;

  version = concatStringsSep "-" (match
    "([[:digit:]]{4})([[:digit:]]{2})([[:digit:]]{2}).*"
    self.lastModifiedDate
  );
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
      ])
      (maybeMissing ../frontend/static/base.css);
  };

  patchPhase = ''
    rm -f frontend/static/base.css
    cp ${css} frontend/static/base.css
  '';
  tags = [ "embed" ];
  modules = ../gomod2nix.toml;
}
