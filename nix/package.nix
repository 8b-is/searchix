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
, testers
}:
let
  inherit (builtins) concatStringsSep match;

  version = "0.1.4";
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
  ldflags = [
    "-s"
    "-w"
    "-X"
    "searchix/internal/config.CommitSHA=${self.rev or self.dirtyRev}"
    "-X"
    "searchix/internal/config.ShortSHA=${self.shortRev or self.dirtyShortRev}"
    "-X"
    "main.buildVersion=${version}"
  ];

  modules = ../gomod2nix.toml;
}
