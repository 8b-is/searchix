package fetcher

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/index"

	"github.com/pkg/errors"
)

type ChannelFetcher struct {
	Source     *config.Source
	SourceFile string
	Logger     *slog.Logger
}

func NewChannelFetcher(
	source *config.Source,
	logger *slog.Logger,
) (*ChannelFetcher, error) {
	switch source.Importer {
	case config.Options:
		return &ChannelFetcher{
			Source: source,
			Logger: logger,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported importer type %s", source.Importer)
	}
}

func (i *ChannelFetcher) FetchIfNeeded(
	ctx context.Context,
	sourceMeta *index.SourceMeta,
) (f FetchedFiles, err error) {
	args := []string{
		"--no-build-output",
		"--timeout",
		strconv.Itoa(int(i.Source.Timeout.Seconds() - 1)),
		fmt.Sprintf("<%s/%s>", i.Source.Channel, i.Source.ImportPath),
		"--attr",
		i.Source.Attribute,
		"--no-out-link",
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

	outPath := path.Join(strings.TrimSpace(string(out)), i.Source.OutputPath, "options.json")
	i.Logger.Debug(
		"checking output path",
		"outputPath",
		outPath,
	)

	if outPath != sourceMeta.Path {
		sourceMeta.Path = outPath
		sourceMeta.Updated = time.Now().Truncate(time.Second)
	}

	file, err := os.Open(outPath)
	if err != nil {
		err = errors.WithMessage(err, "failed to open options.json")

		return
	}

	f = FetchedFiles{
		Options: file,
	}

	return
}
