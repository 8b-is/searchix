{ cfg }:
{ config, lib, name, ... }:
let
  inherit (lib) literalExpression mkOption mkEnableOption types;
in
{
  options = {
    key = mkOption {
      type = types.strMatching "[a-z0-9_-]*";
      default = name;
      description = "URL-safe name for this source.";
    };

    enable = mkEnableOption "Whether to enable this source.";
  };
}
