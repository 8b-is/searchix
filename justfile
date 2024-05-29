default:
	@just --list --justfile {{ justfile() }} --unsorted

prepare:
	nix build .#css -o frontend/static/base.css

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

precommit:
	nix-build -A pre-commit-check

changelog:
	convco changelog | prettier --stdin-filepath CHANGELOG.md > CHANGELOG.md

dev:
	watchexec --no-vcs-ignore --filter "**/*.go" --filter config.toml -r wgo run -exit ./cmd/searchix-web --live --config config.toml

reindex:
	wgo run --exit ./cmd/searchix-web --config config.toml --replace

update:
	wgo run --exit ./cmd/searchix-web --config config.toml --update
