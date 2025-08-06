package importer

import (
	"context"
	"sync"

	"alin.ovh/x/log"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/nix"
	"alin.ovh/searchix/internal/storage"
)

type Processor interface {
	Process(context.Context) (<-chan nix.Importable, <-chan error)
}

func Import[I nix.Importable](
	ctx context.Context,
	log *log.Logger,
	src func(context.Context) (<-chan I, <-chan error),
	dst func(context.Context, <-chan I) <-chan error,
) (bool, error) {
	wg := sync.WaitGroup{}

	wg.Add(1)
	objects, srcErrs := src(ctx)

	wg.Add(1)
	dstErrors := dst(ctx, objects)

	var hadObjectErrors bool
	var criticalError error
	go func() {
		for {
			select {
			case err, running := <-srcErrs:
				if !running {
					wg.Done()
					srcErrs = nil
					log.Debug("processing completed")

					continue
				}
				hadObjectErrors = true
				log.Warn("error processing object from source", "error", err)
			case err, running := <-dstErrors:
				if !running {
					wg.Done()
					dstErrors = nil

					continue
				}
				hadObjectErrors = true
				log.Warn("error writing object to target", "error", err)
			}
		}
	}()

	wg.Wait()

	return hadObjectErrors, criticalError
}

func ImportSource(
	ctx context.Context,
	logger *log.Logger,
	store *storage.Store,
	source *config.Source,
	write *index.WriteIndex,
) (bool, error) {
	var it func(context.Context, *log.Logger, *storage.Store, *config.Source, *index.WriteIndex) (bool, error)
	switch source.Importer {
	case config.Packages:
		it = ImportWithType[nix.Package]
	case config.Options:
		it = ImportWithType[nix.Option]
	}

	return it(ctx, logger, store, source, write)
}

func ImportWithType[I nix.Importable](
	ctx context.Context,
	logger *log.Logger,
	store *storage.Store,
	source *config.Source,
	write *index.WriteIndex,
) (bool, error) {
	return Import(
		ctx,
		logger.Named("importer"),
		storage.MakeSourceExporter[I](store, source),
		func(ctx context.Context, objects <-chan I) <-chan error {
			return write.Import(ctx, nix.ToGenericChannel(objects))
		},
	)
}
