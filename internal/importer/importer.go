package importer

import (
	"context"
	"log/slog"
	"searchix/internal/index"
	"searchix/internal/nix"
	"sync"
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
) bool {
	wg := sync.WaitGroup{}

	wg.Add(1)
	objects, pErrs := processor.Process(ctx)

	wg.Add(1)
	iErrs := indexer.Import(ctx, objects)

	var hadErrors bool
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
				hadErrors = true
				logger.Warn("error ingesting object", "error", err)
			case err, running := <-pErrs:
				if !running {
					wg.Done()
					pErrs = nil
					logger.Debug("processing completed")

					continue
				}
				hadErrors = true
				logger.Warn("error processing object", "error", err)
			}
		}
	}()

	wg.Wait()

	return hadErrors
}
