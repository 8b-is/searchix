package index_test

import (
	"context"
	"maps"
	"math"
	"slices"
	"testing"
	"time"

	"alin.ovh/x/log"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/nix"
	"alin.ovh/searchix/internal/storage"
)

const rootPath = "../../data"

func TestSearchGitPackagesFirst(t *testing.T) {
	log := log.Configure(false)
	cfg := config.DefaultConfig

	root, err := file.CreateAndOpenRoot(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	store, err := storage.New(&storage.Options{
		Root:   root,
		Logger: log.Named("storage"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	read, _, exists, err := index.OpenOrCreate(&index.Options{
		Logger:    log.Named("index"),
		Root:      root,
		BatchSize: cfg.Importer.BatchSize,
		Force:     false,
		LowMemory: false,
		Store:     store,
		Config:    &cfg,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()
	if !exists {
		t.Fatal("expected index to exist")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	source := cfg.Importer.Sources["nixpkgs"]
	if source == nil || !source.Enable {
		t.Fatal("expected source to exist and be enabled")
	}

	result, err := read.Search(
		ctx,
		source,
		"git",
		0,
		100,
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Total < 4 {
		t.Errorf("Expected at least 4 results, got %d", result.Total)
	}
	important := map[string]int{
		"git":        0,
		"git-doc":    0,
		"gitFull":    0,
		"gitMinimal": 0,
		"gitSVN":     0,
	}
	var i int
	for hit := range result.Hits {
		data, ok := hit.Data.(*nix.Package)
		if !ok {
			t.Fatalf("Expected hit.Data to be *nix.Package, got %T", hit.Data)
		}

		if _, found := important[data.Attribute]; found {
			important[data.Attribute] = i
		}
		i++
	}
	if slices.Max(slices.Collect(maps.Values(important))) > len(important) {
		t.Errorf(
			"Expected all of %s to be the first %d matches, got %v",
			slices.Collect(maps.Keys(important)),
			len(important),
			important,
		)
	}
}

func TestSearchJujutsuPackagesFirst(t *testing.T) {
	log := log.Configure(false)
	cfg := config.DefaultConfig

	root, err := file.CreateAndOpenRoot(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	store, err := storage.New(&storage.Options{
		Root:   root,
		Logger: log,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	read, _, exists, err := index.OpenOrCreate(&index.Options{
		Force:     false,
		LowMemory: false,
		Logger:    log.Named("index"),
		Root:      root,
		BatchSize: 0,
		Store:     store,
		Config:    &cfg,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()
	if !exists {
		t.Fatal("expected index to exist")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	source := cfg.Importer.Sources["nixpkgs"]
	if source == nil || !source.Enable {
		t.Fatal("expected source to exist and be enabled")
	}

	result, err := read.Search(
		ctx,
		source,
		"jj",
		0,
		100,
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Total < 4 {
		t.Errorf("Expected at least 4 results, got %d", result.Total)
	}
	important := map[string]int{
		"jj":      0,
		"jujutsu": 0,
		"lazyjj":  0,
		"jjui":    0,
		"jj-fzf":  0,
	}
	matches := []string{}
	unwanted := "javacc"
	unwantedIndex := math.MaxInt
	var i int
	for hit := range result.Hits {
		data, ok := hit.Data.(*nix.Package)
		if !ok {
			t.Fatalf("Expected hit.Data to be *nix.Package, got %T", hit.Data)
		}

		if _, found := important[data.Attribute]; found {
			matches = append(matches, data.Attribute)
		} else if data.Attribute == unwanted {
			unwantedIndex = i
			matches = append(matches, data.Attribute)
		}
		i++
	}
	if slices.Max(slices.Collect(maps.Values(important))) > unwantedIndex {
		t.Errorf(
			"Expected all of %s to be above unwanted result %s at index %d. Results: %v",
			slices.Collect(maps.Keys(important)),
			unwanted,
			unwantedIndex,
			matches,
		)
	}
}
