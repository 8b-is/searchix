# Searchix

A search tool to find options and packages in the NixOS ecosystem. Aims to be like [search.nixos.org](https://search.nixos.org/) with more sources:

-   NixOS options
-   Nix packages
-   [Nix darwin](https://github.com/LnL7/nix-darwin) options
-   [Home manager](https://github.com/nix-community/home-manager) options.

There is an instance running at [searchix.alanpearce.eu](https://searchix.alanpearce.eu/), which uses the following channels, with updates attempted daily:

-   nixos-options: nixos-unstable
-   nixpkgs: nixos-unstable
-   darwin: master
-   home-manager: master

You can also [run it yourself](./docs/running.md), if you're feeling bold. It's very light-weight!

## Status

**Alpha**

Expect breakage. Search results are not expected to match the quality of [search.nixos.org](https://search.nixos.org/), the priority is more on having multiple sources in one location.

Feel free to report bugs at [sourcehut](https://todo.sr.ht/~alanpearce/searchix) or open a discussion either on the [mailing list](https://lists.sr.ht/~alanpearce/searchix-discuss) or on [NixOS community forum](https://discourse.nixos.org/).

## Motivation

I missed having a tool like [search.nixos.org](https://search.nixos.org/) but for [nix darwin](https://github.com/LnL7/nix-darwin) and [home manager](https://github.com/nix-community/home-manager)[^1].

I wanted to contribute something to the nix community. 🩵

Also, I wanted to learn Go!

## Goals

-   Lean/simple
-   Responsive (i.e. low latency)
-   Minimal JavaScript for progressive enhancement
-   Quickly switch between (options) sources with the same query

### Future

-   [ ] `programs.git.enable` should be in the first few results for the query "git"

## Developing

See [docs/developing.md](./docs/developing.md)

[^1]: It was only during development that I found [Home Manager Option Search](https://home-manager-options.extranix.com/)
