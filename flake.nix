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
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    simple-css = {
      url = "github:kevquirk/simple.css";
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
      overlays = {
        default = _: prev: {
          searchix = self.packages.${prev.system}.default;
        };
      };
    }
    // (flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [
            gomod2nix.overlays.default
          ];
        };

        treefmtEval = treefmt-nix.lib.evalModule pkgs {
          projectRootFile = "flake.nix";
          programs = {
            nixfmt.enable = true;
            taplo.enable = true;
            prettier.enable = true;
            sqlfluff.enable = true;
            sqlfluff.dialect = "sqlite";
          };
          settings.formatter = {
            taplo.excludes = [ "gomod2nix.toml" ];
          };
        };
      in
      rec {
        packages.css = pkgs.concatText "simple.css" [ "${simple-css}/simple.css" ];
        packages.default = pkgs.callPackage ./nix/package.nix {
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
          css = "${simple-css}/simple.css";
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
            command = "searchix-web version";
          };
        };
      }
    ));
}
