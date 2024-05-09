default:
	@just --list --justfile {{ justfile() }} --unsorted

prepare:
	ln -sf $(nix-build --no-out-link -A css) frontend/static/base.css

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
	watchexec wgo run -exit ./serve/ --live --config-file config.toml

reindex:
	wgo run -exit ./import/ --replace --config-file config.toml
