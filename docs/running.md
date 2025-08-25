# Running an instance

## Requirements:

- 2GiB disk space
- ~100MiB RAM (requires setting `Importer.LowMemory = true`)

This [repository](https://git.alin.ovh/searchix) (or its [codeberg mirror](https://codeberg.org/alanpearce/searchix)) is importable as a flake containing a NixOS module.

```nix
{
  inputs.searchix.url = "git+https://git.alin.ovh/searchix";
  # or
  inputs.searchix.url = "git+https://codeberg.org/alanpearce/searchix";

  outputs = { self, nixpkgs, searchix }: {
    nixosConfigurations.<hostname> = nixpkgs.lib.nixosSystem {
      modules = [
        searchix.modules.web
      ];
    };
  };
}
```

This will then expose the following schema:

```nix
services.searchix = {
  enable = true;
# user = "searchix";
# group = "searchix";
# homeDir = "/var/lib/searchix";
# settings = {
#   web = {
#     port = 51313;
#     listenAddress = "localhost";
#     baseURL = "http://localhost:51313";
#   };
# };
};
```

Where `settings` complies with [NixOS RFC 42 (settings options)](https://github.com/NixOS/rfcs/pull/42); see the [default values](../defaults.toml) for reference.

## Maintenance

Whilst the service is running, a `SIGUSR1` will cause a complete re-import and a `SIGUSR2` will cause a re-index.
