package programs

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"strings"

	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	_ "modernc.org/sqlite" //nolint:revive // sqlite driver needed for database/sql

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/file"
)

type DB struct {
	source config.Source
	logger *log.Logger
	root   *file.Root
	db     *sql.DB
	stmt   *sql.Stmt
}

type Options struct {
	Logger *log.Logger
	Root   *file.Root
}

func New(source config.Source, options *Options) (*DB, error) {
	db, err := sql.Open("sqlite", fmt.Sprintf(
		"file:%s?mode=%s&_pragma=foreign_keys(1)&_pragma=mmap_size(%d)",
		//nolint:forbidigo // external package
		options.Root.JoinPath(
			source.JoinPath("programs.db"),
		),
		"rwc",
		16*1024*1024,
	))
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to open sqlite database"))
	}
	options.Logger.Debug("opened sqlite database")

	return &DB{
		source: source,

		logger: options.Logger,
		root:   options.Root,
		db:     db,
	}, nil
}

func (p *DB) Instantiate(ctx context.Context) error {
	// nix-instantiate --eval --json -I nixpkgs=channel:nixos-unstable --expr 'toString <nixpkgs/programs.sqlite>'
	args := []string{
		"--eval",
		"--json",
		"-I", fmt.Sprintf("%s=channel:%s", p.source.Key, p.source.Channel),
		"--expr", fmt.Sprintf("toString <%s/%s>", p.source.Key, p.source.Programs.Attribute),
	}

	p.logger.Debug("nix-instantiate command", "args", args)
	cmd := exec.CommandContext(ctx, "nix-instantiate", args...)
	out, err := cmd.Output()
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to run nix-instantiate"))
	}

	outPath := strings.Trim(strings.TrimSpace(string(out)), "\"")
	p.logger.Debug("got output path", "outputPath", outPath)

	_, err = p.db.ExecContext(ctx, "DROP TABLE IF EXISTS programs")
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to drop programs table"))
	}

	_, err = p.db.ExecContext(ctx, "ATTACH DATABASE ? AS input", outPath)
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to attach nix-store programs database"))
	}

	_, err = p.db.ExecContext(ctx, `
CREATE TABLE programs AS
SELECT name, package
FROM input.Programs
GROUP BY name, package
`)
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to create programs table"))
	}
	p.logger.Debug("created programs table")

	_, err = p.db.ExecContext(ctx, `CREATE INDEX idx_package ON programs(package)`)
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to create idx_package index"))
	}
	p.logger.Debug("created idx_package index")

	_, err = p.db.ExecContext(ctx, "DETACH DATABASE input")
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to detach nix-store programs database"))
	}

	return nil
}

func (p *DB) Open(ctx context.Context) (err error) {
	if p.db == nil {
		return fault.New("database not open")
	}

	p.stmt, err = p.db.PrepareContext(ctx, `
SELECT name
FROM programs
WHERE package = ?
`)
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to prepare statement"))
	}
	p.logger.Debug("prepared statement")

	return nil
}

func (p *DB) Close() error {
	if err := p.db.Close(); err != nil {
		return fault.Wrap(err, fmsg.With("failed to close sqlite database"))
	}

	return nil
}

func (p *DB) GetPackagePrograms(ctx context.Context, pkg string) ([]string, error) {
	var programs []string
	rows, err := p.stmt.QueryContext(ctx, pkg)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to execute query"))
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fault.Wrap(err, fmsg.With("failed to scan row"))
		}

		programs = append(programs, name)
	}
	rerr := rows.Close()
	if rerr != nil {
		return nil, fault.Wrap(rerr, fmsg.With("sql error"))
	}

	return programs, nil
}
