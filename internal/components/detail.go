package components

import (
	"go.alanpearce.eu/searchix/internal/nix"

	g "go.alanpearce.eu/gomponents"
)

func Detail(thing nix.Importable) g.Node {
	switch t := thing.(type) {
	case nix.Option:
		return OptionDetail(t)
	case nix.Package:
		return PackageDetail(t)
	default:
		return nil
	}
}

func DetailPage(tdata TemplateData, thing nix.Importable) g.Node {
	return Page(tdata, Detail(thing))
}
