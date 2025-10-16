package importer

import (
	"context"
	"sync"

	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/nix"
)

type Processor interface {
	Process(context.Context) (<-chan nix.Importable, <-chan error)
}

func (imp *Importer) process(
	ctx context.Context,
	processor Processor,
) (bool, error) {
	wg := sync.WaitGroup{}

	wg.Add(1)
	objects, pErrs := processor.Process(ctx)

	wg.Add(1)
	iErrs := imp.options.WriteIndex.Import(ctx, objects)

	var hadObjectErrors bool
	var criticalError error
	go func() {
		for {
			select {
			case err, running := <-iErrs:
				if !running {
					wg.Done()
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
					wg.Done()
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
