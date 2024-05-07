default:
	@just --list --justfile {{ justfile() }} --unsorted

prepare:
	ln -sf $(nix-build --no-out-link -A css) frontend/static/base.css

get-nixpkgs-revision channel="nixos-unstable":
	curl -L https://channels.nixos.org/{{ channel }}/git-revision > data/raw/nixpkgs-{{ channel }}-revision

update-nixpkgs channel="nixos-unstable": (get-nixpkgs-revision channel)
	curl -L https://channels.nixos.org/{{ channel }}/packages.json.br | brotli --stdout --decompress > data/raw/nixpkgs-{{ channel }}.json

update-nixos-options channel="nixos-unstable": (get-nixpkgs-revision channel)
	curl -L https://channels.nixos.org/{{ channel }}/options.json.br | brotli --stdout --decompress > data/raw/nixos-options-{{ channel }}.json

update-darwin-options:
	ln -sf $(nix-build --no-out-link -A darwin-options)/share/doc/darwin/options.json data/raw/darwin-options.json

update-home-manager-options:
	ln -sf $(nix-build --no-out-link -A home-manager-options)/share/doc/home-manager/options.json data/raw/home-manager-options.json

process-nixos-options channel="nixos-unstable":
	wgo run -exit ./process \
		--input data/raw/nixos-options-{{ channel }}.json \
		--output data/processed/nixos-options-{{ channel }}.json \
		--repo NixOS/nixpkgs \
		--revision-file data/raw/nixpkgs-{{ channel }}-revision

process-darwin-options:
	wgo run -exit ./process \
		--input data/raw/darwin-options.json \
		--output data/processed/darwin-options.json \
		--repo LnL7/nix-darwin

process-home-manager-options:
	wgo run -exit ./process \
		--input data/raw/home-manager-options.json \
		--output data/processed/home-manager-options.json \
		--repo nix-community/home-manager

process-all-options: process-nixos-options process-darwin-options process-home-manager-options

update-all-options: update-nixos-options update-darwin-options update-home-manager-options

checkformat:
	gofmt -d .
	goimports -d .

format:
	gofmt -w .
	goimports -w .

fix:
	go fix .

precommit:
	nix-build -A pre-commit-check

dev: prepare
	watchexec wgo run -exit ./serve/ --live
