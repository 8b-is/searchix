package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/yuin/goldmark"

	"alin.ovh/searchix/internal/nixdocs/optlink"
)

func main() {
	// Create a new Goldmark instance with our optlink extension
	markdown := goldmark.New(
		goldmark.WithExtensions(
			optlink.New(),
		),
	)

	// Sample markdown content with option links
	src := []byte(`
# Example using the OptLink Extension

This is a regular paragraph with an empty link that points to an option: [](#opt-systemd.sysusers.enable)

You can use multiple option links in the same document:
- [](#opt-networking.firewall.enable)
- [](#opt-services.openssh.enable)
- [](#opt-users.users.<n>.hashedPassword)

Regular links still work normally: [NixOS](https://nixos.org)
`)

	// Convert the markdown to HTML
	var buf bytes.Buffer
	if err := markdown.Convert(src, &buf); err != nil {
		fmt.Fprintf(os.Stderr, "Error converting markdown: %v\n", err)
		os.Exit(1)
	}

	// Print the resulting HTML
	fmt.Println(buf.String())
}
