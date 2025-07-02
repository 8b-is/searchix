# Searchix

[![Build Status](https://ci.alanpearce.eu/badge/searchix.svg)](https://ci.alanpearce.eu/jobs/searchix)
[![Service Status](https://img.shields.io/endpoint?url=https%3A%2F%2Fstatus.searchix.ovh%2Fshield-badges%2Fstatus.json&style=flat)](https://status.searchix.ovh)

A search tool to find options and packages in the NixOS ecosystem. Aims to be like [search.nixos.org](https://search.nixos.org/) with more sources:

- NixOS options
- Nix packages
- [Nix darwin](https://github.com/LnL7/nix-darwin) options
- [Home manager](https://github.com/nix-community/home-manager) options
- [Nix User Repository](https://github.com/nix-community/NUR) packages

There is an instance running at [searchix.ovh](https://searchix.ovh/), which uses the following channels, with updates attempted daily:

- nixos-options: nixos-unstable
- nixpkgs: nixos-unstable
- darwin: master branch
- home-manager: master branch
- nur: main branch

You can also [run it yourself](./docs/running.md), if you're feeling bold. It's quite lightweight!

## Status

**Beta**

Expect breakage. Search results are not expected to match the quality of [search.nixos.org](https://search.nixos.org/), the priority is more on having multiple sources in one location.

Feel free to report bugs at [codeberg](https://codeberg.org/alanpearce/searchix/issues).

## Motivation

I missed having a tool like [search.nixos.org](https://search.nixos.org/) but for [nix darwin](https://github.com/LnL7/nix-darwin) and [home manager](https://github.com/nix-community/home-manager)[^1].

I wanted to contribute something to the nix community. 🩵

Also, I wanted to learn Go!

## Goals

- Lean/simple
- Responsive (i.e. low latency)
- Minimal JavaScript for progressive enhancement
- Quickly switch between (options) sources with the same query
- Querying "git" should give relevant results:
    - Options: programs.git.{enable,package}
    - Packages: git/gitMinimal/gitFull...

### Future

## Developing

See [docs/developing.md](./docs/developing.md)

[^1]: It was only during development that I found [Home Manager Option Search](https://home-manager-options.extranix.com/)
