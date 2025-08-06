package importer

import (
	"testing"

	"alin.ovh/x/log"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/manpages"
	"alin.ovh/searchix/internal/storage"
)

var cfg = config.DefaultConfig

func BenchmarkImporterLowMemory(b *testing.B) {
	tmp, err := file.OpenRoot(b.TempDir())
	if err != nil {
		b.Fatal(err)
	}

	logger := log.Configure(false)
	_, write, _, err := index.OpenOrCreate(&index.Options{
		Config:    &cfg,
		Force:     false,
		LowMemory: true,
		BatchSize: cfg.Importer.BatchSize,
		Logger:    logger.Named("index"),
		Root:      tmp,
	})
	if err != nil {
		b.Fatal(err)
	}

	store, err := storage.New(&storage.Options{
		Root:   tmp,
		Logger: logger.Named("storage"),
	})
	if err != nil {
		b.Fatal(err)
	}

	imp, err := New(&cfg, &Options{
		Logger:     logger.Named("importer"),
		LowMemory:  true,
		WriteIndex: write,
		Manpages: manpages.New(&manpages.Options{
			Logger: logger.Named("manpages"),
			Root:   tmp,
		}),
		Offline: true,
		Root:    tmp,
		Storage: store,
	})
	if err != nil {
		b.Fatal(err)
	}

	source := cfg.Importer.Sources["nixpkgs"]
	source.Programs.Enable = false

	err = imp.Start(
		b.Context(),
		false,
		false,
		&[]string{source.Key},
	)
	if err != nil {
		b.Fatal(err)
	}

	hadErrors, err := ImportSource(
		b.Context(),
		logger.Named("importer"),
		store,
		source,
		write,
	)
	if err != nil {
		b.Fatal(err)
	}
	if hadErrors {
		b.Fatal("had errors")
	}
}
