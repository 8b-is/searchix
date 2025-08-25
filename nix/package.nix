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
}:
let
  version = "0.2.4";
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
    "-s"
    "-w"
    "-X"
    "alin.ovh/searchix/internal/config.Version=${version}"
  ];

  postInstall = ''
    install -d -m755 $out/lib/searchix/static
    cp -r $src/frontend/static $out/lib/searchix
    cp ${css} $out/lib/searchix/static/base.css
    $out/bin/searchix-web generate-error-page > $out/lib/searchix/error.html
  '';

  modules = ../gomod2nix.toml;

  doCheck = false;
}
