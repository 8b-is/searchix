package importer

import (
	"log/slog"
	"searchix/internal/config"
	"searchix/internal/index"
	"testing"
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
