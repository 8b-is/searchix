package manpages_test

import (
	"testing"

	"alin.ovh/x/log"

	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/manpages"
)

func TestURLMapOpenAndFallback(t *testing.T) {
	logger := log.Configure(false)
	tmpDir := t.TempDir()

	root, err := file.CreateAndOpenRoot(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	err = root.WriteFile("manpage-urls.json", []byte(`{
		"configuration.nix(5)": "https://nixos.org/manual/nixos/stable/#ch-configuration",
		"nix.conf(5)": "https://nixos.org/manual/nix/stable/command-ref/conf-file.html"
	}`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	err = root.WriteFile("manpage-urls-home-manager.json", []byte(`{
		"home-configuration.nix(5)": "https://nix-community.github.io/home-manager/options.xhtml"
	}`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	mdb := manpages.New(&manpages.Options{
		Logger: logger.Named("manpages"),
		Root:   root,
	})
	if err := mdb.Open(); err != nil {
		t.Fatalf("Open() failed: %v", err)
	}

	t.Run("mapped nixpkgs manpage", func(t *testing.T) {
		got, ok := mdb.Get("5", "configuration.nix")
		if !ok {
			t.Fatal("expected configuration.nix(5) to be found")
		}
		want := "https://nixos.org/manual/nixos/stable/#ch-configuration"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("mapped home-manager manpage", func(t *testing.T) {
		got, ok := mdb.Get("5", "home-configuration.nix")
		if !ok {
			t.Fatal("expected home-configuration.nix(5) to be found")
		}
		want := "https://nix-community.github.io/home-manager/options.xhtml"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("fallback to mankier for unmapped manpage (wofi(5))", func(t *testing.T) {
		got, ok := mdb.Get("5", "wofi")
		if !ok {
			t.Fatal("expected wofi(5) fallback to succeed")
		}
		want := "https://www.mankier.com/5/wofi"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("unsectioned lookup", func(t *testing.T) {
		got, ok := mdb.Get("", "nix.conf")
		if !ok {
			t.Fatal("expected nix.conf without section to resolve")
		}
		want := "https://nixos.org/manual/nix/stable/command-ref/conf-file.html"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}

		gotFallback, ok := mdb.Get("", "wofi")
		if !ok {
			t.Fatal("expected wofi without section to fall back to mankier")
		}
		if gotFallback != "https://www.mankier.com/wofi" {
			t.Errorf("got %q, want https://www.mankier.com/wofi", gotFallback)
		}
	})
}
