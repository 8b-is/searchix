default:
	@just --list --justfile {{ justfile() }} --unsorted

prepare:
	nix build .#css -o frontend/static/base.css

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

dev:
	watchexec --no-vcs-ignore --filter "**/*.go" --filter config.toml -r wgo run -exit ./ --live --config config.toml

reindex:
	wgo run --exit . --config config.toml --replace

update:
	wgo run --exit . --config config.toml --update
