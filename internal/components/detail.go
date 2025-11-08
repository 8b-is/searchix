package components

import (
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/nix"

	g "alin.ovh/gomponents"
)

func Detail(thing index.Indexable) g.Node {
	switch t := thing.(type) {
	case nix.Option:
		return OptionDetail(t)
	case nix.Package:
		return PackageDetail(t)
	default:
		return nil
	}
}

func DetailPage(tdata TemplateData, thing index.Indexable) g.Node {
	return Page(tdata, Detail(thing))
}
