package components_test

import (
	"bytes"
	"strings"
	"testing"

	"alin.ovh/searchix/internal/components"
	"alin.ovh/searchix/internal/nix"
)

func TestPackageDetailTryCommandsAndVersionHistory(t *testing.T) {
	pkg := nix.Package{
		Name:        "ripgrep",
		Attribute:   "ripgrep",
		Version:     "14.1.1",
		Description: "Line-oriented search tool",
		MainProgram: "rg",
	}

	var buf bytes.Buffer
	if err := components.PackageDetail(pkg).Render(&buf); err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	html := buf.String()

	expectedSnippets := []string{
		"<dt>Try</dt>",
		"<code>nix-shell -p ripgrep</code>",
		"<code>nix run nixpkgs#ripgrep</code>",
		"<dt>Version</dt>",
		"14.1.1",
		"https://lazamar.co.uk/nix-versions/?channel=nixpkgs-unstable&amp;package=ripgrep",
		"https://www.nixhub.io/packages/ripgrep",
	}

	for _, snippet := range expectedSnippets {
		if !strings.Contains(html, snippet) {
			t.Errorf("expected rendered PackageDetail HTML to contain %q, got:\n%s", snippet, html)
		}
	}
}
