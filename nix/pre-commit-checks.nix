{ pkgs }:
rec {
  src = ../.;
  hooks = {
    gotest.enable = true;
    golangci-lint.enable = true;
    staticcheck =
      let
        wrapper = pkgs.symlinkJoin {
          name = "staticcheck-wrapped";
          paths = [ pkgs.go-tools ];
          nativeBuildInputs = [ pkgs.makeWrapper ];
          postBuild = ''
            wrapProgram $out/bin/staticcheck \
              --prefix PATH : ${pkgs.lib.makeBinPath [ pkgs.go ]}
          '';
        };
      in
      {
        enable = true;
        package = wrapper;
        entry =
          let
            script = pkgs.writeShellScript "precommit-staticcheck" ''
              err=0
              for dir in $(echo "$@" | xargs -n1 dirname | sort -u); do
                ${hooks.staticcheck.package}/bin/staticcheck ./"$dir"
                code="$?"
                if [[ "$err" -eq 0 ]]; then
                  err="$code"
                fi
              done
              exit $err
            '';
          in
          builtins.toString script;
      };

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
      types_or = [ "plain-text" "yaml" "gotmpl" "javascript" ];
      settings = {
        plugins = with pkgs.nodePackages; [
          "${prettier-plugin-go-template}/lib/node_modules/prettier-plugin-go-template/lib/index.js"
        ];
      };
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
            ${pkgs.wgo}/bin/wgo run -exit searchix.go --print-default-config > defaults.toml
          '';
        in
        builtins.toString script;
    };
  };
}
