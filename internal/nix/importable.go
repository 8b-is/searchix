package nix

import (
	"encoding/gob"

	"alin.ovh/searchix/internal/config"
)

type Importable interface {
	ImporterType() string
	BleveType() string
	GetName() string
	GetSource() string
}

func GetKey(i Importable) string {
	return i.ImporterType() + "/" + i.GetSource() + "/" + i.GetName()
}

func MakeKey(source config.Source, id string) string {
	return source.Importer.String() + "/" + source.Key + "/" + id
}

func init() {
	gob.Register(Option{})
	gob.Register(Package{})
}
