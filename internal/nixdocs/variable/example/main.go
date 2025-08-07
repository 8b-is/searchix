package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/yuin/goldmark"

	"alin.ovh/searchix/internal/nixdocs/variable"
)

func main() {
	// Create a new Goldmark instance with our variable extension
	md := goldmark.New(
		goldmark.WithExtensions(
			variable.New(),
		),
	)

	// Some example Markdown content with variable references
	src := []byte(`
# Variable References Example

Nix packages like {var}` + "`pkgs.virtualbox`" + ` and {var}` + "`pkgs.firefox`" + `
can be installed in your system.

The {var}` + "`services.postgresql.enable`" + ` option enables the PostgreSQL service.

Configuration variables are often found in these contexts:
- {var}` + "`system.stateVersion`" + ` for system configuration
- {var}` + "`networking.hostName`" + ` for network settings
- {var}` + "`users.users.alice`" + ` for user accounts
`)

	// Convert the markdown to HTML
	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		fmt.Println("Conversion error:", err)
		os.Exit(1)
	}

	// Print the HTML output
	fmt.Println(buf.String())
}
