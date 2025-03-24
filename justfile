default:
	@just --list --justfile {{ justfile() }} --unsorted

prepare:
	cp -f $(nix build --print-out-paths .#css) frontend/static/base.css

update-go-dependencies:
	go get -u all
	go mod tidy
	gomod2nix

update-nix-dependencies:
	nix flake update

update-dependencies: update-go-dependencies update-nix-dependencies

checkformat:
	gofmt -d .
	goimports -d .
	golines --dry-run .

format:
	gofmt -w .
	goimports -w .
	golines -w .

fix:
	go fix .

build:
	nix build .#

generate-defaults:
	wgo run --exit ./cmd/searchix-web --print-default-config > defaults.toml

precommit:
	nix-build -A pre-commit-check

clean-workdir:
	git diff --no-ext-diff --quiet

changelog:
	convco changelog -u $(convco version --bump) | \
		sd -F HEAD $(convco version --print-prefix --bump) | \
		prettier --stdin-filepath CHANGELOG.md > CHANGELOG.md

update-nix-package-version:
	#!/usr/bin/env bash
	VER=$(convco version --bump)
	sd 'version = "[^"]+"' "version = \"$VER\"" nix/package.nix
	sd 'h=v.+$' "h=v$VER" nix/modules/default.nix

release: clean-workdir changelog update-nix-package-version
	git add CHANGELOG.md nix/package.nix nix/modules/default.nix
	git commit -m "chore: release $(convco version --bump)"
	git tag $(convco version --bump --print-prefix)

run: dev

dev:
	modd

reindex *flags:
	wgo run --exit ./cmd/searchix-web --config config.toml --replace --dev {{ flags }}
