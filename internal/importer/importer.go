package importer

import (
	"context"
	"log/slog"
	"sync"

	"go.alanpearce.eu/searchix/internal/index"
	"go.alanpearce.eu/searchix/internal/nix"
)

type Importer interface {
	Import(context.Context, *index.WriteIndex) (bool, error)
}

type Processor interface {
	Process(context.Context) (<-chan nix.Importable, <-chan error)
}

func process(
	ctx context.Context,
	indexer *index.WriteIndex,
	processor Processor,
	logger *slog.Logger,
) (bool, error) {
	wg := sync.WaitGroup{}

	wg.Add(1)
	objects, pErrs := processor.Process(ctx)

	wg.Add(1)
	iErrs := indexer.Import(ctx, objects)

	var hadObjectErrors bool
	var criticalError error
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
