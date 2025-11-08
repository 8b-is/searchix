package nix

import (
	"encoding/gob"

	"alin.ovh/searchix/internal/config"
)

func MakeKey(source config.Source, id string) string {
	return source.Importer.String() + "/" + source.Key + "/" + id
}

func init() {
	gob.Register(Option{})
	gob.Register(Package{})
}
