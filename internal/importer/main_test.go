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
		Logger:    logger.Named("index"),
	})
	if err != nil {
		b.Fatal(err)
	}

	imp := New(&cfg, logger.Named("importer"), write)
	err = imp.Start(
		context.Background(),
		false,
		&[]string{"nixpkgs"},
	)
	if err != nil {
		b.Fatal(err)
	}
}
