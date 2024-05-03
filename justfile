default:
	@just --list --justfile {{ justfile() }} --unsorted

prepare:
	ln -sf $(nix-build --no-out-link -A css) frontend/static/base.css

update-nixos-options:
	ln -sf $(nix-build --no-out-link -A nixos-options)/share/doc/nixos/options.json data/nixos-options.json

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
