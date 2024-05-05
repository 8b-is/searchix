{ nixpkgs ? <nixpkgs>
, pkgs ? import nixpkgs { }
, system ? builtins.currentSystem
, stateVersion ? pkgs.lib.version
, ...
}:
let
  inherit (pkgs) lib;
  inherit (lib) hasPrefix removePrefix;

  nixos = pkgs.nixos ({ lib, ... }: {
    nixpkgs.hostPlatform = system;
    system.stateVersion = lib.versions.majorMinor stateVersion;
  });

  inherit (nixos.config.system.nixos) revision;

  gitHubDeclaration = user: repo: ref: subpath:
    # Default to `master` if we don't know what revision the system
    # configuration is using (custom nixpkgs, etc.).
    let urlRef = if ref != null then ref else "master";
    in {
      url = "https://github.com/${user}/${repo}/blob/${urlRef}/${subpath}";
      name = "<${repo}/${subpath}>";
    };

  doc = pkgs.nixosOptionsDoc {
    inherit (nixos) options;
    transformOptions = opt: opt // {
      declarations =
        map
          (decl:
            if hasPrefix (toString nixpkgs) (toString decl)
            then
              gitHubDeclaration "NixOS" "nixpkgs" revision
                (removePrefix "/"
                  (removePrefix (toString nixpkgs) (toString decl)))
            else if decl == "lib/modules.nix" then
              gitHubDeclaration "NixOS" "nixpkgs" revision decl
            else decl)
          opt.declarations;
    };
  };
in
doc.optionsJSON
