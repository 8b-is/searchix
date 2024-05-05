default:
	@just --list --justfile {{ justfile() }} --unsorted

prepare:
	ln -sf $(nix-build --no-out-link -A css) frontend/static/base.css

update-nixos-options:
	wgo run -exit ./process --input $(nix-build --no-out-link -A nixos-options)/share/doc/nixos/options.json --output data/nixos-options.json

update-darwin-options:
	wgo run -exit ./process --input $(nix-build --no-out-link -A darwin-options)/share/doc/darwin/options.json --output data/darwin-options.json

update-home-manager-options:
	wgo run -exit ./process --input $(nix-build --no-out-link -A home-manager-options)/share/doc/home-manager/options.json --output data/home-manager-options.json

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
	wgo run ./serve/ --live
