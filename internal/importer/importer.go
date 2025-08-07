package importer

import (
	"context"
	"sync"

	"alin.ovh/x/log"

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

func pipe(
	ctx context.Context,
	log *log.Logger,
	src ImportSource,
	dst ImportDestination,
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
