package main

import (
	"github.com/blevesearch/bleve/v2/cmd/bleve/cmd"

	// to support standard set of build tags
	_ "github.com/blevesearch/bleve/v2/config"

	// support nixAttr
	_ "alin.ovh/searchix/internal/index/nixattr"
)

func main() {
	cmd.Execute()
}
