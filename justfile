default:
	@just --list --justfile {{ justfile() }} --unsorted

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

dev:
	wgo run ./serve/ --live
