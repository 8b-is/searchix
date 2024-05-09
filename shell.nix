{ pkgs ? (
    let
      sources = import ./nix/sources.nix;
    in
    import sources.nixpkgs {
      overlays = [
        (import "${sources.gomod2nix}/overlay.nix")
        (import ./nix/overlays/default.nix)
      ];
    }
  )
}:

let
  goEnv = pkgs.mkGoEnv { pwd = ./.; };
  searchix = import ./default.nix {
    inherit pkgs;
  };
in
pkgs.mkShell {
  packages = with pkgs; [
    goEnv

    brotli
    bleve
    wgo
    gomod2nix
    niv
    nixpkgs-lint

    bun
  ];
  shellHook = ''
    ${searchix.pre-commit-check.shellHook}
  '';
  buildInputs = searchix.pre-commit-check.enabledPackages;
}
