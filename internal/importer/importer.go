package importer

import (
	"context"
	"sync"

	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/nix"
	"alin.ovh/x/log"
	"gitlab.com/tozd/go/errors"
)

type Processor interface {
	Process(context.Context) (<-chan nix.Importable, <-chan errors.E)
}

func process(
	ctx context.Context,
	indexer *index.WriteIndex,
	processor Processor,
	logger *log.Logger,
) (bool, errors.E) {
	wg := sync.WaitGroup{}

	wg.Add(1)
	objects, pErrs := processor.Process(ctx)

	wg.Add(1)
	iErrs := indexer.Import(ctx, objects)

	var hadObjectErrors bool
	var criticalError errors.E
	go func() {
		for {
			select {
			case err, running := <-iErrs:
				if !running {
					wg.Done()
					iErrs = nil
					logger.Debug("ingest completed")

					continue
				}
				be, isBatchError := err.(*index.BatchError)
				if isBatchError {
					criticalError = be

					break
				}
				hadObjectErrors = true
				logger.Warn("error ingesting object", "error", err)
			case err, running := <-pErrs:
				if !running {
					wg.Done()
					pErrs = nil
					logger.Debug("processing completed")

					continue
				}
				hadObjectErrors = true
				logger.Warn("error processing object", "error", err)
			}
		}
	}()

	wg.Wait()

	return hadObjectErrors, criticalError
}
