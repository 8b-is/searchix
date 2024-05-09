self: super: {
  nodePackages = super.nodePackages.extend
    (final: prev:
      super.callPackage ./prettier-plugin-go-template { }
    );
  bleve = super.callPackage ./bleve.nix { };
}
