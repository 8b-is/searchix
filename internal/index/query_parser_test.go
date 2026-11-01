package index_test

import (
	"context"
	"testing"
	"time"

	"alin.ovh/x/log"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/nix"
	"alin.ovh/searchix/internal/nixdocs"
)

func TestAdvancedQuerySyntax(t *testing.T) {
	logger := log.Configure(false)
	cfg := config.DefaultConfig

	tmpDir := t.TempDir()
	root, err := file.CreateAndOpenRoot(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	read, write, err := index.OpenOrCreate(&index.Options{
		Logger:    logger.Named("index"),
		Root:      root,
		BatchSize: 100,
		Force:     true,
		LowMemory: false,
		Config:    &cfg,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()

	now := time.Now()
	docs := []index.Indexable{
		nix.Option{
			Name:        "systemd.user.services",
			Source:      "home-manager",
			Description: nixdocs.Markdown("Definition of systemd per-user service units."),
			Loc:         []string{"systemd", "user", "services"},
			ImportedAt:  now,
		},
		nix.Option{
			Name:        "systemd.services",
			Source:      "nixos",
			Description: nixdocs.Markdown("Definition of systemd system service units."),
			Loc:         []string{"systemd", "services"},
			ImportedAt:  now,
		},
		nix.Package{
			Name:        "systemd-manager",
			Attribute:   "systemd-manager",
			Source:      "nixpkgs",
			Version:     "1.4.2",
			Description: "TUI tool to manage systemd service units",
			MainProgram: "systemd-manager",
			Programs:    []string{"systemd-manager"},
			Platforms:   []string{"x86_64-linux"},
			ImportedAt:  now,
		},
		nix.Package{
			Name:        "git",
			Attribute:   "git",
			Source:      "nixpkgs",
			Version:     "2.47.0",
			Description: "Distributed version control system",
			MainProgram: "git",
			Programs:    []string{"git", "git-shell"},
			Platforms:   []string{"x86_64-linux", "aarch64-darwin"},
			ImportedAt:  now,
		},
		nix.Package{
			Name:        "git-svn",
			Attribute:   "gitSVN",
			Source:      "nixpkgs",
			Version:     "2.47.0",
			Description: "Subversion bridge for git distributed version control",
			MainProgram: "git-svn",
			Programs:    []string{"git-svn"},
			Platforms:   []string{"x86_64-linux"},
			ImportedAt:  now,
		},
	}

	err = write.WithBatch(func(batch *index.Batcher) {
		for _, item := range docs {
			doc, mapErr := write.MapDocument(item)
			if mapErr != nil {
				t.Fatalf("MapDocument failed: %v", mapErr)
			}
			if idxErr := batch.IndexAdvanced(doc); idxErr != nil {
				t.Fatalf("IndexAdvanced failed: %v", idxErr)
			}
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	allSource := config.Source{
		Key:      "all",
		Importer: config.All,
	}

	t.Run("parenthesized OR with implicit AND ((home-manager OR nixpkgs) systemd service)", func(t *testing.T) {
		res, err := read.Search(ctx, index.SearchRequest{
			Source:   allSource,
			Keyword:  "(home-manager OR nixpkgs) systemd service",
			PageSize: 50,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Total != 2 {
			t.Fatalf("expected 2 results (home-manager and nixpkgs, excluding nixos), got %d", res.Total)
		}
		for hit := range res.Hits {
			if hit.Data.GetSource() == "nixos" {
				t.Errorf("did not expect nixos document in (home-manager OR nixpkgs) query, got %v", hit.ID)
			}
		}
	})

	t.Run("case-insensitive field prefix and negation (name:git -svn)", func(t *testing.T) {
		res, err := read.Search(ctx, index.SearchRequest{
			Source:   allSource,
			Keyword:  "name:git -svn",
			PageSize: 50,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Total != 1 {
			t.Fatalf("expected 1 result for 'name:git -svn', got %d", res.Total)
		}
		for hit := range res.Hits {
			if hit.Data.GetName() != "git" {
				t.Errorf("expected 'git', got %q", hit.Data.GetName())
			}
		}
	})

	t.Run("NOT operator (systemd service NOT nixos)", func(t *testing.T) {
		res, err := read.Search(ctx, index.SearchRequest{
			Source:   allSource,
			Keyword:  "systemd service NOT source:nixos",
			PageSize: 50,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Total != 2 {
			t.Fatalf("expected 2 results, got %d", res.Total)
		}
	})

	t.Run("version field search (version:2.47.0)", func(t *testing.T) {
		res, err := read.Search(ctx, index.SearchRequest{
			Source:   allSource,
			Keyword:  "version:2.47.0",
			PageSize: 50,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Total != 2 {
			t.Fatalf("expected 2 results with version 2.47.0, got %d", res.Total)
		}
	})
}
