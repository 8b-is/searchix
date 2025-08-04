package importer

import (
	"context"
	"sync"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/nix"
)

type Processor interface {
	Process(context.Context) (<-chan nix.Importable, <-chan error)
}

func (imp *Importer) process(
	ctx context.Context,
	processor Processor,
	source *config.Source,
) (bool, error) {
	wg := sync.WaitGroup{}

	wg.Add(1)
	objects, pErrs := processor.Process(ctx)

	d1, d2 := duplicate(objects)

	wg.Add(1)
	iErrs := imp.options.WriteIndex.Import(ctx, d1)

	wg.Add(1)
	wErrs := imp.options.Storage.Import(ctx, source, d2)

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
					imp.options.Logger.Debug("processing completed")

					continue
				}
				hadObjectErrors = true
				imp.options.Logger.Warn("error processing object", "error", err)
			case err, running := <-wErrs:
				if !running {
					wg.Done()
					wErrs = nil

					continue
				}
				hadObjectErrors = true
				imp.options.Logger.Warn("error writing to storage", "error", err)
			}
		}
	}()

	wg.Wait()

	return hadObjectErrors, criticalError
}

func duplicate[T any](v <-chan T) (<-chan T, <-chan T) {
	if v == nil {
		return nil, nil
	}

	dup1 := make(chan T, 1)
	dup2 := make(chan T, 1)

	go func() {
		for v := range v {
			dup1 <- v
			dup2 <- v
		}
		close(dup1)
		close(dup2)
	}()

	return dup1, dup2
}
