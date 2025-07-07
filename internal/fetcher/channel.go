package fetcher

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
)

type ChannelFetcher struct {
	Source     *config.Source
	SourceFile string
	Logger     *log.Logger
}

func NewChannelFetcher(
	source *config.Source,
	logger *log.Logger,
) (*ChannelFetcher, error) {
	switch source.Importer {
	case config.Options:
		return &ChannelFetcher{
			Source: source,
			Logger: logger,
		}, nil
	default:
		return nil, fault.Newf("unsupported importer type %s", source.Importer)
	}
}

func (i *ChannelFetcher) FetchIfNeeded(
	ctx context.Context,
	sourceMeta *index.SourceMeta,
) (*FetchedFiles, error) {
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
	out, err := cmd.Output()
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to run nix-build (--dry-run)"))
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
		return nil, fault.Wrap(err, fmsg.With("failed to open options.json"))
	}

	return &FetchedFiles{
		Options: file,
	}, nil
}
