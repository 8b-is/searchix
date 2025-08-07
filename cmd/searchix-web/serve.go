package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"

	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/importer"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/manpages"
	"alin.ovh/searchix/internal/server"
	"alin.ovh/searchix/internal/storage"
	"alin.ovh/searchix/web"
)

type ServeOptions struct{}

func (opts *ServeOptions) Execute(_ []string) (err error) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	root, err := file.CreateAndOpenRoot(cfg.DataPath)
	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to open data root"))
	}
	defer root.Close()

	store, err := storage.New(&storage.Options{
		Root:   root,
		Logger: logger.Named("store"),
	})
	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to create store"))
	}
	defer store.Close()

	read, write, exists, err := index.OpenOrCreate(
		&index.Options{
			Config:    cfg,
			LowMemory: cfg.Importer.LowMemory,
			BatchSize: cfg.Importer.BatchSize,
			Logger:    logger.Named("index"),
			Root:      root,
			Store:     store,
		},
	)
	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to open or create index"))
	}

	mdb := manpages.New(&manpages.Options{
		Logger: logger.Named("manpages"),
		Root:   root,
	})

	s, err := web.New(cfg, logger, &server.Options{
		ReadIndex:      read,
		ManpagesURLMap: mdb,
		Store:          store,
	})
	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to initialise searchix-web"))
	}

	imp, err := importer.New(cfg, &importer.Options{
		Storage:    store,
		WriteIndex: write,
		LowMemory:  cfg.Importer.LowMemory,
		Logger:     logger.Named("importer"),
		Manpages:   mdb,
		Root:       root,
	})
	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to create importer"))
	}

	if !exists {
		err = imp.Start(ctx, true, false, nil)
		if err != nil {
			return fault.Wrap(err, fmsg.With("Failed to start importer"))
		}

		for _, source := range cfg.Importer.Sources {
			hadErrors, err := imp.IndexSource(ctx, source)
			if err != nil {
				return fault.Wrap(err, fmsg.Withf("Failed to import source %s", source.Name))
			}

			if hadErrors {
				logger.Warn("Imported source encountered errors", "source", source.Name)
			}
		}
	}

	err = imp.EnsureSourcesIndexed(ctx, read)
	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to ensure sources indexed"))
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		sCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		defer wg.Done()
		err := s.Start(sCtx, globalOptions.Dev)
		if err != nil && !errors.Is(err, context.Canceled) {
			// Error starting or closing listener:
			logger.Fatal("error", "error", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		imp.StartUpdateTimer(ctx)
	}()

	<-ctx.Done()
	s.Stop()
	wg.Wait()

	return
}

func init() {
	var opts ServeOptions

	cmd, err := parser.AddCommand("serve", "run server", "Serve web interface", &opts)
	if err != nil {
		panic(err)
	}

	cmd.Aliases = []string{"run"}
}
