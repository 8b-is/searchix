package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"

	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/importer"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/manpages"
	"alin.ovh/searchix/internal/storage"
)

type IngestOptions struct {
	Fetch   bool `long:"fetch"   description:"pre-fetch data"`
	Offline bool `long:"offline" description:"offline mode"`
	Replace bool `long:"replace" description:"replace existing storage"`
	Reindex bool `long:"reindex" description:"reindex existing index"`
}

func (opts *IngestOptions) Execute(_ []string) error {
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

	_, write, exists, err := index.OpenOrCreate(
		&index.Options{
			Config:    cfg,
			Force:     opts.Reindex,
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

	imp, err := importer.New(cfg, &importer.Options{
		Storage:    store,
		WriteIndex: write,
		LowMemory:  cfg.Importer.LowMemory,
		Logger:     logger.Named("importer"),
		Manpages:   mdb,
		Root:       root,
		Offline:    opts.Offline,
	})
	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to create importer"))
	}

	if !exists || opts.Replace || opts.Fetch {
		importer.MarkImportStarted()

		err = imp.Fetch(ctx, true, opts.Fetch && !opts.Replace, nil)
		if err != nil {
			return fault.Wrap(err, fmsg.With("Failed to start importer"))
		}

		importer.MarkImportFinished()
		importer.MarkLastRun(write.Meta)
		err = write.SaveMeta()
		if err != nil {
			return fault.Wrap(err, fmsg.With("Failed to save index metadata"))
		}
	}

	if !exists || opts.Reindex {
		err = imp.Index(ctx)
		if err != nil {
			return fault.Wrap(err, fmsg.With("Failed to index data"))
		}
	}

	if opts.Replace || opts.Reindex {
		err = imp.Prune(ctx)
		if err != nil {
			return fault.Wrap(err, fmsg.With("Failed to prune index"))
		}
	}

	return nil
}

func init() {
	var opts IngestOptions

	cmd, err := parser.AddCommand(
		"ingest",
		"Ingest data",
		"Fetch, store and index data",
		&opts,
	)
	if err != nil {
		panic(err)
	}

	cmd.Aliases = []string{"import"}
}
