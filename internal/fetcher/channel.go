package fetcher

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"searchix/internal/config"
	"searchix/internal/file"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type ChannelFetcher struct {
	DataPath   string
	Source     *config.Source
	SourceFile string
	Logger     *slog.Logger
}

func NewChannelFetcher(
	source *config.Source,
	dataPath string,
	logger *slog.Logger,
) (*ChannelFetcher, error) {
	switch source.Importer {
	case config.Options:
		return &ChannelFetcher{
			DataPath: dataPath,
			Source:   source,
			Logger:   logger,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported importer type %s", source.Importer)
	}
}

func (i *ChannelFetcher) FetchIfNeeded(
	parent context.Context,
) (f FetchedFiles, updated bool, err error) {
	ctx, cancel := context.WithTimeout(parent, i.Source.FetchTimeout.Duration)
	defer cancel()

	dest := i.DataPath

	var before string
	before, err = os.Readlink(dest)
	if file.NeedNotExist(err) != nil {
		err = errors.WithMessagef(err, "could not call readlink on file %s", dest)

		return
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

	if i.Source.URL != "" {
		args = append(args, "-I", fmt.Sprintf("%s=%s", i.Source.Channel, i.Source.URL))
	}

	i.Logger.Debug("nix-build command", "args", args)
	cmd := exec.CommandContext(ctx, "nix-build", args...)
	var out []byte
	out, err = cmd.Output()
	if err != nil {
		err = errors.WithMessage(err, "failed to run nix-build (--dry-run)")

		return
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
	var after string
	after, err = os.Readlink(dest)
	if err = file.NeedNotExist(err); err != nil {
		err = errors.WithMessagef(
			err,
			"failed to stat output file from nix-build, filename: %s",
			outPath,
		)

		return
	}
	i.Logger.Debug("stat after", "name", after)

	updated = before != after

	f = FetchedFiles{
		Options: path.Join(dest, i.Source.OutputPath, "options.json"),
	}

	return
}
