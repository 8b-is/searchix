package importer

import (
	"testing"

	"alin.ovh/x/log"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/manpages"
)

var cfg = config.DefaultConfig

func BenchmarkImporterLowMemory(b *testing.B) {
	b.ReportAllocs()

	tmp, err := file.OpenRoot(b.TempDir())
	if err != nil {
		b.Fatal(err)
	}

	logger := log.Configure(false)
	_, write, err := index.OpenOrCreate(&index.Options{
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

	imp, err := New(&cfg, &Options{
		Logger:     logger.Named("importer"),
		LowMemory:  true,
		WriteIndex: write,
		Manpages: manpages.New(&manpages.Options{
			Logger: logger.Named("manpages"),
			Root:   tmp,
		}),
		Offline: false,
		Root:    tmp,
	})
	if err != nil {
		b.Fatal(err)
	}

	source := cfg.Importer.Sources["nixpkgs"]
	source.Programs.Enable = false

	err = imp.Fetch(
		b.Context(),
		false,
		[]string{source.Key},
	)
	if err != nil {
		b.Fatal(err)
	}
}
