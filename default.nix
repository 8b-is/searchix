let
  sources = import ./nix/sources.nix;
in
{ pkgs ? (
    import sources.nixpkgs {
      overlays = [
        (import "${sources.gomod2nix}/overlay.nix")
        (import ./nix/overlays/default.nix)
      ];
    }
  )
}:
{
  css = pkgs.fetchurl {
    inherit (sources.simple-css) url sha256;
  };

  nixos-options = import (./. + "/importers/nixos-options.nix") {
    inherit pkgs;
  };

  searchix = pkgs.buildGoApplication {
    pname = "searchix";
    version = "0.1";
    pwd = ./.;
    src = ./.;
    modules = ./gomod2nix.toml;
  };

  pre-commit-check = (import sources.git-hooks).run {
    src = ./.;
    hooks = {
      gotest.enable = true;
      golangci-lint.enable = true;
      staticcheck.enable = true;

      statix = {
        enable = true;
        settings.ignore = [ "nix/*" ];
      };
      deadnix = {
        enable = true;
        excludes = [ "nix/.*" ];
      };
      nixpkgs-fmt = {
        enable = true;
        excludes = [ "nix/.*" ];
      };

      convco.enable = true;

      check-case-conflicts.enable = true;
      check-symlinks.enable = true;
      editorconfig-checker.enable = true;
      prettier = {
        enable = true;
        types_or = [ "plain-text" "yaml" "gotmpl" ];
        settings = {
          plugins = with pkgs.nodePackages; [
            "${prettier-plugin-go-template}/lib/node_modules/prettier-plugin-go-template/lib/index.js"
          ];
        };
      };

      yamllint = {
        enable = true;
      };
      go-mod-tidy = {
        enable = true;
        name = "go-mod-tidy";
        description = "Run `go mod tidy`";
        types_or = [ "go" "go-mod" ];
        entry = "${pkgs.go}/bin/go mod tidy";
        pass_filenames = false;
      };
      gomod2nix = {
        enable = true;
        name = "gomod2nix";
        description = "Import go.mod updates to nix";
        types_or = [ "go-sum" ];
        entry = "${pkgs.gomod2nix}/bin/gomod2nix";
        pass_filenames = false;
      };
    };
  };
}
