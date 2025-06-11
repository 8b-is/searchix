package main

import (
	"bytes"
	"fmt"
	"os"

	"go.alanpearce.eu/searchix/internal/nixdocs/envvar"

	"github.com/yuin/goldmark"
)

func main() {
	// Create a new markdown parser with the environment variable extension
	md := goldmark.New(
		goldmark.WithExtensions(
			envvar.New(),
		),
	)

	// Example markdown text with environment variable references
	markdown := "You need to set {env}`XDG_DATA_DIRS` to include your custom data directories.\n\n" +
		"Other important environment variables are {env}`HOME` and {env}`PATH`."

	// Convert markdown to HTML
	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		fmt.Fprintf(os.Stderr, "Error converting markdown: %v\n", err)
		os.Exit(1)
	}

	// Print the resulting HTML
	fmt.Println("HTML Output:")
	fmt.Println("===========")
	fmt.Println(buf.String())
}
