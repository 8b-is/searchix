{ cfg, settingsFormat }:
{ config, lib, name, ... }:
let
  inherit (lib) literalExpression mkOption mkEnableOption types;
in
{
  freeformType = settingsFormat.type;

  options = {
    key = mkOption {
      type = types.strMatching "[a-z0-9_-]*";
      default = name;
      description = "URL-safe name for this source.";
    };

    enable = mkEnableOption "Whether to enable this source.";
  };
}
