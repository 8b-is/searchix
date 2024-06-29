package importer

import (
	"log/slog"
	"testing"

	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/index"
)

var cfg = config.DefaultConfig

func BenchmarkImporterLowMemory(b *testing.B) {
	tmp := b.TempDir()
	cfg.LogLevel = slog.LevelDebug
	_, write, _, err := index.OpenOrCreate(tmp, false, &index.Options{
		LowMemory: true,
	})
	if err != nil {
		b.Fatal(err)
	}

	err = Start(&cfg, write, false, &[]string{"nixpkgs"})
	if err != nil {
		b.Fatal(err)
	}
}
