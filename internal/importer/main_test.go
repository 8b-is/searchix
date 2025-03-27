package importer

import (
	"context"
	"testing"

	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/index"
	"go.alanpearce.eu/x/log"
)

var cfg = config.DefaultConfig

func BenchmarkImporterLowMemory(b *testing.B) {
	tmp := b.TempDir()
	logger := log.Configure(false)
	_, write, _, err := index.OpenOrCreate(tmp, false, &index.Options{
		LowMemory: true,
		BatchSize: cfg.Importer.BatchSize,
		Logger:    logger.Named("index"),
	})
	if err != nil {
		b.Fatal(err)
	}

	imp, err := New(&cfg, &Options{
		Logger:     logger.Named("importer"),
		LowMemory:  true,
		WriteIndex: write,
	})
	if err != nil {
		b.Fatal(err)
	}

	err = imp.Start(
		context.Background(),
		false,
		&[]string{"nixpkgs"},
	)
	if err != nil {
		b.Fatal(err)
	}
}
