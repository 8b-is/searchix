package nix

import (
	"encoding/gob"

	"go.alanpearce.eu/searchix/internal/config"
)

type Importable interface {
	BleveType() string
	GetName() string
	GetSource() string
}

func GetKey(i Importable) string {
	return i.BleveType() + "/" + i.GetSource() + "/" + i.GetName()
}

func MakeKey(source *config.Source, id string) string {
	return source.Importer.Singular() + "/" + source.Key + "/" + id
}

func init() {
	gob.Register(Option{})
	gob.Register(Package{})
}
