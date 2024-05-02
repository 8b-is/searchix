# Searchix

## Development

### Pre-requisites

-   [Nix (the package manager)](https://nixos.org/download/)
-   Optional:
    -   [cachix](https://docs.cachix.org/installation#installation)
    -   [lorri](https://github.com/nix-community/lorri)
    -   [direnv](https://direnv.net/)
    
### Setup

1. Optional: use cachix to avoid building some tools

    ```shell-session
    cachix use pre-commit-hooks
    ```

2. Start a nix shell (this will take some time)

    ```shell-session
    nix-shell
    ```

3. Optional: install and set up lorri and use `lorri shell` instead of `nix-shell`

4. Optional: install direnv, integrate it with your shell and then run

    ```shell-session
    ln -s .envrc.sample .envrc
    ```

    Now `nix-shell` or `lorri shell` will be run for you automatically when you `cd` into this directory.
