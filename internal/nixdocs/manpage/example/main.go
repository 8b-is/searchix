package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/yuin/goldmark"
	"go.alanpearce.eu/searchix/internal/nixdocs/manpage"
)

func main() {
	// Create a new Goldmark instance with our manpage extension
	md := goldmark.New(
		goldmark.WithExtensions(
			manpage.New(),
		),
	)

	// Some example Markdown content with manpage references
	src := []byte(`
# Man Page References Example

Basic command documentation can be found in {manpage}` + "`bash(1)`" + ` and {manpage}` + "`zsh(1)`" + `.

System configuration is documented in {manpage}` + "`nix.conf(5)`" + `.

Common system administration commands:
- {manpage}` + "`systemctl(1)`" + ` for managing services
- {manpage}` + "`journalctl(1)`" + ` for viewing logs
- {manpage}` + "`mount(8)`" + ` for filesystem operations

See {manpage}` + "`man(1)`" + ` for information about the manual pages system itself.
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
