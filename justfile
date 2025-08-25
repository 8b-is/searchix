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
	golangci-lint fmt --diff

format:
	golangci-lint fmt

fix:
	go fix .

test:
	go test ./...

check-flake:
	nix flake check

ci: checkformat test

build:
	nix build .#

generate-defaults:
	go run ./cmd/searchix-web defaults > defaults.toml

check-config:
	go run ./cmd/searchix-web check-config --config config.toml

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
	sd '/v[0-9.]+/' "/v$VER/" nix/modules/default.nix

release: clean-workdir changelog update-nix-package-version
	#!/bin/sh
	set -eu
	VERSION=$(convco version --bump --print-prefix)
	git add CHANGELOG.md nix/package.nix nix/modules/default.nix
	git commit -m "chore: release ${VERSION}"
	git tag "${VERSION}"
	git push origin main "${VERSION}"

run: dev

dev:
	watchexec --restart --exts go,toml,mod -- go run ./cmd/searchix-web --config config.toml serve --dev

ingest *flags:
	go run ./cmd/searchix-web --config config.toml ingest --dev {{ flags }}

fetch: (ingest "--fetch")

replace: (ingest "--replace")

reindex: (ingest "--reindex")
