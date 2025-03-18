{ pkgs }:
rec {
  src = ../.;
  hooks = {
    gotest.enable = false;
    golangci-lint.enable = true;

    staticcheck = {
      enable = true;
    };
    statix = {
      enable = true;
      # ignore is a glob
      settings.ignore = [ "nix/overlays/*" ];
    };
    deadnix = {
      enable = true;
      # excludes is a regex
      excludes = [ "nix/overlays/.*" ];
    };
    nixpkgs-fmt = {
      enable = true;
      excludes = [ "nix/overlays/.*" ];
    };

    convco.enable = true;

    check-case-conflicts.enable = true;
    check-symlinks.enable = true;
    editorconfig-checker = {
      enable = true;
      excludes = [ "\.md$" "\.ya?ml$" ];
    };
    prettier = {
      enable = true;
      types_or = [ "plain-text" "yaml" "javascript" ];
      excludes = [ "CHANGELOG\.md$" ];
    };

    yamllint = {
      enable = true;
    };
    golines = {
      enable = true;
      name = "golines";
      description = "A golang formatter that fixes long lines";
      types_or = [ "go" ];
      entry = "${pkgs.golines}/bin/golines --write-output";
      pass_filenames = true;
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

    generate-default-config = {
      enable = true;
      name = "generate-default-config";
      description = "Ensure default configuration file is up-to-date";
      files = "^defaults.toml|internal/config/";
      types_or = [ "go" "toml" ];
      entry =
        let
          script = pkgs.writeShellScript "generate-default-config" ''
            ${pkgs.wgo}/bin/wgo run -exit ./cmd/searchix-web --print-default-config > defaults.toml
          '';
        in
        builtins.toString script;
    };
  };
}
