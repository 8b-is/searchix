package nix

import (
	"encoding/gob"

	"alin.ovh/searchix/internal/config"
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

func ToGenericChannel[I Importable](in <-chan I) <-chan Importable {
	out := make(chan Importable)
	go func() {
		for i := range in {
			out <- i
		}
		close(out)
	}()

	return out
}

func init() {
	gob.Register(Option{})
	gob.Register(Package{})
}
