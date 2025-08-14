{
  pre-commit-check,
  mkGoEnv,
  gomod2nix,
  mkShell,
  pkgs,
}:
let
  goEnv = mkGoEnv { pwd = ../.; };
in
mkShell {
  packages =
    with pkgs;
    [
      goEnv

      sd
      watchexec
      brotli
      boltbrowser
      bleve
      gomod2nix
    ]
    ++ pre-commit-check.enabledPackages;
  inherit (pre-commit-check) shellHook;
}
