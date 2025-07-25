package fetcher

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/index"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
)

type ChannelFetcher struct {
	Source     *config.Source
	SourceFile string
	*Options
}

func NewChannelFetcher(
	source *config.Source,
	options *Options,
) (*ChannelFetcher, error) {
	switch source.Importer {
	case config.Options:
		return &ChannelFetcher{
			Source:  source,
			Options: options,
		}, nil
	default:
		return nil, fault.Newf("unsupported importer type %s", source.Importer)
	}
}

func (i *ChannelFetcher) FetchIfNeeded(
	ctx context.Context,
	sourceMeta *index.SourceMeta,
) (*FetchedFiles, error) {
	target := i.Source.JoinPath("options.json")
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

	//nolint:forbidigo // nix-build only gives the top-level path
	outPath := filepath.Join(strings.TrimSpace(string(out)), i.Source.OutputPath, "options.json")

	i.Logger.Debug(
		"checking output path",
		"outputPath",
		outPath,
		"updated",
		outPath != sourceMeta.Path,
	)
	if outPath != sourceMeta.Path {
		sourceMeta.Path = outPath
		sourceMeta.Updated = time.Now().Truncate(time.Second)

		if exists, err := i.Root.Exists(target); err != nil {
			return nil, fault.Wrap(err, fmsg.With("failed to check if target path exists"))
		} else if exists {
			err := i.Root.Remove(target)
			if err != nil {
				return nil, fault.Wrap(err, fmsg.Withf("failed to remove previous file %s", target))
			}
		}

		err := i.Root.CopyTo(outPath, target)
		if err != nil {
			return nil, fault.Wrap(err, fmsg.With("failed to copy options.json"))
		}
	}

	file, err := i.Root.Open(target)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to open options.json"))
	}

	return &FetchedFiles{
		Options: file,
	}, nil
}
