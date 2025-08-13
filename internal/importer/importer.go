package importer

import (
	"context"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/nix"
	"alin.ovh/searchix/internal/storage"
)

type Processor interface {
	Process(context.Context) (<-chan nix.Importable, <-chan error)
}

type (
	ImportSource      func(context.Context) (<-chan nix.Importable, <-chan error)
	ImportDestination func(context.Context, <-chan nix.Importable) <-chan error
)

func (imp *Importer) indexSource(
	ctx context.Context,
	source *config.Source,
) (bool, error) {
	var exporter func(context.Context) (<-chan nix.Importable, <-chan error)
	switch source.Importer {
	case config.Packages:
		exporter = storage.MakeSourceExporter[nix.Package](imp.options.Storage, source)
	case config.Options:
		exporter = storage.MakeSourceExporter[nix.Option](imp.options.Storage, source)
	}

	return pipe(
		ctx,
		imp.options.Logger,
		exporter,
		func(ctx context.Context, objects <-chan nix.Importable) <-chan error {
			return imp.options.WriteIndex.Import(ctx, objects)
		},
	)
}
