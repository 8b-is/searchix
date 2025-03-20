package programs

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"strings"

	"gitlab.com/tozd/go/errors"
	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/x/log"
	_ "modernc.org/sqlite" //nolint:blank-imports // sqlite driver needed for database/sql
)

type DB struct {
	Path   string
	Source *config.Source

	logger *log.Logger
	db     *sql.DB
	stmt   *sql.Stmt
}

func Instantiate(
	ctx context.Context,
	source *config.Source,
	logger *log.Logger,
) (*DB, errors.E) {
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

func (p *DB) Open() errors.E {
	var err error
	p.db, err = sql.Open("sqlite", p.Path)
	if err != nil {
		return errors.WithMessage(err, "failed to open sqlite database")
	}
	p.logger.Debug("opened sqlite database")

	_, err = p.db.Exec("ATTACH DATABASE ':memory:' AS mem")
	if err != nil {
		return errors.WithMessage(err, "failed to attach in-memory database")
	}

	_, err = p.db.Exec(`
CREATE TABLE mem.programs AS
SELECT name, package
FROM main.Programs
GROUP BY name, package
`)
	if err != nil {
		return errors.WithMessage(err, "failed to create programs table")
	}
	p.logger.Debug("created programs table")

	_, err = p.db.Exec(`CREATE INDEX mem.idx_package ON programs(package)`)
	if err != nil {
		return errors.WithMessage(err, "failed to create idx_package index")
	}
	p.logger.Debug("created idx_package index")

	p.stmt, err = p.db.Prepare(`
SELECT name
FROM mem.programs
WHERE package = ?
`)
	if err != nil {
		return errors.WithMessage(err, "failed to prepare statement")
	}
	p.logger.Debug("prepared statement")

	return nil
}

func (p *DB) Close() errors.E {
	if err := p.db.Close(); err != nil {
		return errors.WithMessage(err, "failed to close sqlite database")
	}

	return nil
}

func (p *DB) GetPackagePrograms(ctx context.Context, pkg string) ([]string, errors.E) {
	var programs []string
	if p.db == nil {
		return nil, errors.New("database not open")
	}
	rows, err := p.stmt.QueryContext(ctx, pkg)
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

	return programs, nil
}
