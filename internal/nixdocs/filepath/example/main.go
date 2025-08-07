package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/yuin/goldmark"

	"alin.ovh/searchix/internal/nixdocs/filepath"
)

func main() {
	// Create a new Goldmark instance with our file path extension
	md := goldmark.New(
		goldmark.WithExtensions(
			filepath.New(),
		),
	)

	// Some example Markdown content with file path references
	src := []byte(`
# File Path References Example

System configuration files like {file}` + "`/etc/passwd`" + ` and {file}` + "`/etc/shadow`" +
		` contain user account information.

The {file}` + "`/etc/fstab`" + ` file defines the file systems to be mounted.

Configuration files are often found in the following locations:
- {file}` + "`/etc/`" + ` for system-wide configuration
- {file}` + "`~/.config/`" + ` for user-specific configuration
- {file}` + "`/usr/local/etc/`" + ` for locally installed software
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
