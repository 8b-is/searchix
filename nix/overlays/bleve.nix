{
  lib,
  stdenv,
  fetchFromGitHub,
  installShellFiles,
  buildGoModule,
}:
let
  gomod = builtins.fromTOML (builtins.readFile ./../../gomod2nix.toml);
  version = gomod.mod."github.com/blevesearch/bleve/v2".version;
in
buildGoModule rec {
  pname = "bleve";
  inherit version;

  src = fetchFromGitHub {
    owner = "blevesearch";
    repo = "bleve";
    rev = version;
    hash = "sha256-QycDP5IRuZsLo0cDbsQoX4Qb4yjsV3PDns96+dMajoE=";
  };

  vendorHash = "sha256-RRIDViTRRKSN6QJhUSvTCoXJ2R5E3DWT4ygk9U/Q1dA=";

  subPackages = [ "cmd/bleve" ];

  nativeBuildInputs = [ installShellFiles ];
  postInstall = lib.optionalString (stdenv.buildPlatform == stdenv.targetPlatform) ''
    installShellCompletion --cmd bleve \
      --bash <($out/bin/bleve completion bash) \
      --fish <($out/bin/bleve completion fish) \
      --zsh <($out/bin/bleve completion zsh)
  '';

  meta = {
    description = "Command-line tool to interact with bleve indexes";
    homepage = "http://blevesearch.com";
    licenses = lib.licenses.asl20;
  };
}
