{ config
, lib
, pkgs
, ...
}:

let
  cfg = config.services.searchix;

  package = (import ../.. { inherit pkgs; }).searchix;

  settingsFormat = pkgs.formats.toml { };

  env = {
    ENVIRONMENT = "production";
    LISTEN_ADDRESS = cfg.listenAddress;
    PORT = (toString cfg.port);
    BASE_URL = cfg.baseUrl;
    CONFIG_FILE = settingsFormat.generate "searchix-config.toml" cfg.settings;
    LOG_LEVEL = cfg.logLevel;
  };

  defaultServiceConfig = {
    User = cfg.user;
    Group = cfg.group;
    ReadWritePaths = [ cfg.homeDir ];
    StateDirectory = mkIf (cfg.homeDir == "/var/lib/searchix") [ "searchix" ];
    Restart = "on-failure";

    CacheDirectory = "searchix";
    CapabilityBoundingSet = "";
    DeviceAllow = "";
    LockPersonality = true;
    MemoryDenyWriteExecute = true;
    NoNewPrivileges = true;
    PrivateDevices = true;
    PrivateMounts = true;
    PrivateTmp = true;
    PrivateUsers = true;
    ProtectClock = true;
    ProtectHome = true;
    ProtectHostname = true;
    ProtectSystem = "strict";
    ProtectControlGroups = true;
    ProtectKernelLogs = true;
    ProtectKernelModules = true;
    ProtectKernelTunables = true;
    RestrictAddressFamilies = [ "AF_UNIX" "AF_INET" "AF_INET6" ];
    RestrictNamespaces = true;
    RestrictRealtime = true;
    RestrictSUIDSGID = true;
    SystemCallArchitectures = "native";
    SystemCallFilter = [ "@system-service" "~@privileged @setuid @keyring" ];
    UMask = "0066";
  };

  inherit (lib) mkEnableOption mkOption mkIf optionalAttrs types;
in
{
  options.services.searchix = {
    enable = mkEnableOption "Searchix options search";

    user = mkOption {
      type = types.str;
      default = "searchix";
      description = "User account under which searchix runs.";
    };

    group = mkOption {
      type = types.str;
      default = "searchix";
      description = "Group under which searchix runs.";
    };

    homeDir = mkOption {
      type = types.path;
      default = "/var/lib/searchix";
      description = "Home directory for searchix user";
    };

    dates = mkOption {
      type = types.singleLineStr;
      default = "04:00";
      example = "weekly";
    };

    port = mkOption {
      type = types.port;
      description = "Port for searchix to listen on";
      default = 51313;
    };

    listenAddress = mkOption {
      type = types.str;
      description = "Listen on a specific IP address.";
      default = "localhost";
    };

    baseUrl = mkOption {
      type = types.str;
      description = "The base URL that searchix will be served on.";
      default = "http://localhost:3000";
    };

    sentryDsn = mkOption {
      type = with types; nullOr str;
      description = "Optionally enable sentry to track errors.";
      default = null;
    };

    logLevel = mkOption {
      type = with types; enum [ "error" "warn" "info" "debug" ];
      description = "Only log messages with the given severity or above.";
      default = "info";
    };

    importTimeout = mkOption {
      type = types.str;
      default = "30m";
      description = ''
        Maximum time to wait for all import jobs.
        May need to be increased based on the number of sources.
      '';
    };

    settings = mkOption {
      type = types.submodule {
        freeformType = settingsFormat.type;
        options = {
          data-path = mkOption {
            type = types.str;
            description = "Where to store search index and other data, can be relative to homeDir.";
            default = "${cfg.homeDir}/data";
          };
          sources = mkOption {
            type = with types; attrsOf (submodule (import ./source-options.nix { inherit cfg; }));
            default = {
              nixos.enable = true;
              darwin.enable = false;
              home-manager.enable = false;
            };
            description = "Declarative specification of options sources for searchix.";
          };
        };
      };
      default = { };
      description = "Configuration for searchix (TODO: publish description).";
    };
  };

  config = mkIf cfg.enable {
    nixpkgs.overlays = [
      (import "${(import ../sources.nix).gomod2nix}/overlay.nix")
    ];

    systemd.services.searchix-importer = {
      description = "Searchix option importer";
      unitConfig.Conflicts = [ "searchix-web" ];
      path = with pkgs; [ nix ];
      serviceConfig = defaultServiceConfig // {
        ExecStart = "${package}/bin/import";
        Type = "oneshot";
      };
      environment = env;
      startAt = cfg.dates;
    };

    systemd.timers.searchix-importer = {
      timerConfig = {
        Persistent = true;
        RandomizedDelaySec = 1800;
      };
    };

    systemd.services.searchix-web = {
      description = "Searchix Nix option search";
      after = [ "searchix-importer.service" ];
      wantedBy = [ "multi-user.target" ];
      environment = env;
      serviceConfig = defaultServiceConfig // {
        ExecStart = "${package}/bin/serve";
      } // lib.optionalAttrs (cfg.port < 1024) {
        AmbientCapabilities = [ "CAP_NET_BIND_SERVICE" ];
        CapabilityBoundingSet = [ "CAP_NET_BIND_SERVICE" ];
      };
    };

    users.users = optionalAttrs (cfg.user == "searchix") {
      searchix = {
        group = cfg.group;
        home = cfg.homeDir;
        isSystemUser = true;
      };
    };

    users.groups = optionalAttrs (cfg.group == "searchix") {
      searchix = { };
    };
  };
}
