package importer

import (
	"context"
	"sync"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"

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
	errs := make(chan error)
	wg.Go(func() {
		processor.Process(ctx, objects, errs)
		close(objects)
	})

	wg.Go(func() {
		err := imp.options.WriteIndex.Import(ctx, objects, errs)
		if err != nil {
			criticalError = fault.Wrap(err, fmsg.With("error writing batch"))
		}
		close(errs)
	})

	for err := range errs {
		hadObjectErrors = true
		imp.options.Logger.Warn("error processing object", "error", err)
	}

	wg.Wait()
	imp.options.Logger.Debug("ingest completed")

	return hadObjectErrors, criticalError
}
