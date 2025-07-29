package importer

import (
	"context"
	"testing"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/manpages"
	"alin.ovh/x/log"
)

var cfg = config.DefaultConfig

func BenchmarkImporterLowMemory(b *testing.B) {
	tmp, err := file.OpenRoot(b.TempDir())
	if err != nil {
		b.Fatal(err)
	}

	logger := log.Configure(false)
	_, write, _, err := index.OpenOrCreate(&index.Options{
		Force:     false,
		LowMemory: true,
		BatchSize: cfg.Importer.BatchSize,
		Logger:    logger.Named("index"),
		Root:      tmp,
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
	})
	if err != nil {
		b.Fatal(err)
	}

	err = imp.Start(
		context.Background(),
		false,
		false,
		&[]string{"nixpkgs"},
	)
	if err != nil {
		b.Fatal(err)
	}
}
