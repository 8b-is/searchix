package main

import (
	"bytes"
	"fmt"
	"os"

	"alin.ovh/searchix/internal/nixdocs/option"

	"github.com/yuin/goldmark"
)

func main() {
	// Create a new markdown parser with the option extension
	md := goldmark.New(
		goldmark.WithExtensions(
			option.New(),
		),
	)

	// Example markdown text with option references
	markdown := "Use {option}`initialHashedPassword` to set the initial hashed password for a user."

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
