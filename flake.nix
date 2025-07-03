{
  description = "A basic gomod2nix flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    pre-commit-hooks = {
      url = "github:cachix/pre-commit-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs = {
        nixpkgs.follows = "nixpkgs";
        flake-utils.follows = "flake-utils";
      };
    };
    treefmt-nix.url = "github:numtide/treefmt-nix";
    simple-css = {
      url = "https://raw.githubusercontent.com/kevquirk/simple.css/v2.3.4/simple.css";
      flake = false;
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      gomod2nix,
      pre-commit-hooks,
      treefmt-nix,
      simple-css,
    }:
    {
      nixosModules = {
        web = import ./nix/modules self;
      };
    }
    // (flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [
            (import ./nix/overlays)
            gomod2nix.overlays.default
          ];
        };

        treefmtEval = treefmt-nix.lib.evalModule pkgs {
          projectRootFile = "flake.nix";
          programs = {
            nixfmt.enable = true;
            taplo.enable = true;
            prettier.enable = true;
            gofumpt.enable = true;
            golines.enable = true;
            sqlfluff.enable = true;
            sqlfluff.dialect = "sqlite";
          };
          settings.formatter = {
            taplo.excludes = [ "gomod2nix.toml" ];
          };
        };
      in
      rec {
        packages.css = "${simple-css}";
        packages.default = pkgs.callPackage ./nix/package.nix {
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
          css = simple-css;
        };
        formatter = treefmtEval.config.build.wrapper;
        devShells.default = pkgs.callPackage ./nix/dev-shell.nix {
          pre-commit-check = {
            inherit (self.checks.${system}.pre-commit-check)
              shellHook
              enabledPackages
              ;
          };
          inherit (gomod2nix.legacyPackages.${system}) mkGoEnv gomod2nix;
        };
        checks = {
          pre-commit-check = pre-commit-hooks.lib.${system}.run (
            import ./nix/pre-commit-checks.nix {
              inherit pkgs treefmtEval;
            }
          );
          buildVersion = pkgs.testers.testVersion {
            package = packages.default;
          };
        };
      }
    ));
}
