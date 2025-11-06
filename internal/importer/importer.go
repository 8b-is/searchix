package importer

import (
	"context"
	"sync"

	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/nix"
)

type Processor interface {
	Process(context.Context, chan<- nix.Importable, chan<- error)
}

func (imp *Importer) process(
	ctx context.Context,
	processor Processor,
) (hadObjectErrors bool, criticalError error) {
	wg := sync.WaitGroup{}

	objects := make(chan nix.Importable, 1)
	pErrs := make(chan error)
	wg.Go(func() {
		processor.Process(ctx, objects, pErrs)
	})

	iErrs := make(chan error)
	wg.Go(func() {
		imp.options.WriteIndex.Import(ctx, objects, iErrs)
	})

	go func() {
		for {
			select {
			case err, running := <-iErrs:
				if !running {
					iErrs = nil
					imp.options.Logger.Debug("ingest completed")

					continue
				}
				be, isBatchError := err.(*index.BatchError)
				if isBatchError {
					criticalError = be

					break
				}
				hadObjectErrors = true
				imp.options.Logger.Warn("error ingesting object", "error", err)
			case err, running := <-pErrs:
				if !running {
					pErrs = nil

					continue
				}
				hadObjectErrors = true
				imp.options.Logger.Warn("error processing object", "error", err)
			}
		}
	}()

	wg.Wait()

	return hadObjectErrors, criticalError
}
