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
	return pipe(
		ctx,
		imp.options.Logger,
		storage.MakeSourceExporter(imp.options.Storage, source),
		func(ctx context.Context, objects <-chan nix.Importable) <-chan error {
			return imp.options.WriteIndex.Import(ctx, objects)
		},
	)
}
