package importer

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"searchix/internal/file"
	"searchix/internal/search"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type ChannelImporter struct {
	DataPath   string
	Source     *Source
	SourceFile string
	Logger     *slog.Logger
}

func (i *ChannelImporter) FetchIfNeeded(parent context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(parent, i.Source.FetchTimeout)
	defer cancel()

	dest := i.DataPath

	before, err := os.Readlink(dest)
	if file.NeedNotExist(err) != nil {
		return false, errors.WithMessagef(err, "could not call readlink on file %s", dest)
	}
	i.Logger.Debug("stat before", "name", before)

	args := []string{
		"--no-build-output",
		"--timeout",
		strconv.Itoa(int(i.Source.FetchTimeout.Seconds() - 1)),
		fmt.Sprintf("<%s/%s>", i.Source.Channel, i.Source.ImportPath),
		"--attr",
		i.Source.Attribute,
		"--out-link",
		dest,
	}

	i.Logger.Debug("nix-build command", "args", args)
	cmd := exec.CommandContext(ctx, "nix-build", args...)
	out, err := cmd.Output()
	if err != nil {
		return false, errors.WithMessage(err, "failed to run nix-build (--dry-run)")
	}
	i.Logger.Debug("nix-build", "output", strings.TrimSpace(string(out)))

	outPath := path.Join(dest, i.Source.OutputPath)
	i.Logger.Debug(
		"checking output path",
		"outputPath",
		outPath,
		"dest",
		dest,
		"source",
		i.Source.OutputPath,
	)
	after, err := os.Readlink(dest)
	if err := file.NeedNotExist(err); err != nil {
		return false, errors.WithMessagef(
			err,
			"failed to stat output file from nix-build, filename: %s",
			outPath,
		)
	}
	i.Logger.Debug("stat after", "name", after)

	return before != after, nil
}

func (i *ChannelImporter) Import(parent context.Context, indexer *search.WriteIndex) (bool, error) {
	if i.Source.OutputPath == "" {
		return false, errors.New("no output path specified")
	}

	filename := path.Join(i.DataPath, i.SourceFile, i.Source.OutputPath)
	i.Logger.Debug("preparing import run", "revision", i.Source.Repo.Revision, "filename", filename)

	return processOptions(parent, indexer, &importConfig{
		Source:   i.Source,
		Filename: filename,
		Logger:   i.Logger,
	})
}
