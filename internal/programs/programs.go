package programs

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/x/log"
	_ "modernc.org/sqlite" //nolint:blank-imports // sqlite driver needed for database/sql
)

type DB struct {
	Path   string
	Source *config.Source

	logger *log.Logger
	db     *sql.DB
}

func Instantiate(ctx context.Context, source *config.Source, logger *log.Logger) (*DB, error) {
	// nix-instantiate --eval --json -I nixpkgs=channel:nixos-unstable --expr 'toString <nixpkgs/programs.sqlite>'
	args := []string{
		"--eval",
		"--json",
		"-I", fmt.Sprintf("%s=channel:%s", source.Key, source.Channel),
		"--expr", fmt.Sprintf("toString <%s/%s>", source.Key, source.Programs.Attribute),
	}

	logger.Debug("nix-instantiate command", "args", args)
	cmd := exec.CommandContext(ctx, "nix-instantiate", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to run nix-instantiate")
	}

	outPath := strings.Trim(strings.TrimSpace(string(out)), "\"")
	logger.Debug("got output path", "outputPath", outPath)

	return &DB{
		Source: source,
		Path:   outPath,

		logger: logger,
	}, nil
}

func (p *DB) Open() error {
	db, err := sql.Open("sqlite", p.Path)
	if err != nil {
		return errors.WithMessage(err, "failed to open sqlite database")
	}
	p.db = db

	return nil
}

func (p *DB) Close() error {
	if err := p.db.Close(); err != nil {
		return errors.WithMessage(err, "failed to close sqlite database")
	}

	return nil
}

func (p *DB) GetPackagePrograms(ctx context.Context, pkg string) (programs []string, err error) {
	if p.db == nil {
		return nil, errors.New("database not open")
	}
	rows, err := p.db.QueryContext(ctx, `
SELECT name
FROM Programs
WHERE package = ?
GROUP BY name, package`, pkg)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to execute query")
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, errors.WithMessage(err, "failed to scan row")
		}

		programs = append(programs, name)
	}
	rerr := rows.Close()
	if rerr != nil {
		return nil, errors.WithMessage(rerr, "sql error")
	}

	return
}
