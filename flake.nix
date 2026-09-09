{
  description = "Search tool for options and packages in the NixOS ecosystem";

  inputs = {
    nixpkgs.url = "https://channels.nixos.org/nixpkgs-unstable/nixexprs.tar.zst";
    pre-commit-hooks = {
      url = "github:cachix/pre-commit-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs = {
        nixpkgs.follows = "nixpkgs";
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
      gomod2nix,
      pre-commit-hooks,
      treefmt-nix,
      simple-css,
    }:
    let
      supportedSystems = [
        "aarch64-darwin"
        "aarch64-linux"
        "x86_64-linux"
      ];

      forEachSupportedSystem =
        f:
        nixpkgs.lib.genAttrs supportedSystems (
          system:
          f {
            inherit system;
            pkgs = import nixpkgs {
              inherit system;
              overlays = [
                gomod2nix.overlays.default
              ];
            };
          }
        );

      treefmtEvalFor =
        pkgs:
        treefmt-nix.lib.evalModule pkgs {
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
    {
      nixosModules = {
        web = import ./nix/modules self;
      };
      overlays = {
        default = _: prev: {
          searchix = self.packages.${prev.system}.default;
        };
      };

      packages = forEachSupportedSystem (
        { pkgs, system }:
        {
          css = pkgs.concatText "simple.css" [ "${simple-css}/simple.css" ];
          default = pkgs.callPackage ./nix/package.nix {
            inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
            css = "${simple-css}/simple.css";
          };
        }
      );

      formatter = forEachSupportedSystem ({ pkgs, ... }: (treefmtEvalFor pkgs).config.build.wrapper);

      devShells = forEachSupportedSystem (
        { pkgs, system }:
        {
          default = pkgs.callPackage ./nix/dev-shell.nix {
            pre-commit-check = {
              inherit (self.checks.${system}.pre-commit-check)
                shellHook
                enabledPackages
                ;
            };
            inherit (gomod2nix.legacyPackages.${system}) mkGoEnv gomod2nix;
          };
        }
      );

      checks = forEachSupportedSystem (
        { pkgs, system, ... }:
        {
          pre-commit-check = pre-commit-hooks.lib.${system}.run (
            import ./nix/pre-commit-checks.nix {
              inherit pkgs;
              treefmtEval = treefmtEvalFor pkgs;
            }
          );
          buildVersion = pkgs.testers.testVersion {
            package = self.packages.${system}.default;
            command = "searchix-web version";
          };
        }
      );
    };
}
