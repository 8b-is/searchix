package components

import (
	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/nix"

	g "alin.ovh/gomponents"
	. "alin.ovh/gomponents/html"
)

func Packages(result *index.Result) g.Node {
	return Table(
		THead(
			Tr(
				Th(Scope("col"), g.Text("Attribute")),
				Th(Scope("col"), g.Text("Name")),
				Th(Scope("col"), Class("description"), g.Text("Description")),
				g.If(config.DevMode,
					Th(Scope("col"), Class("score"), g.Text("Score")),
				),
			),
		),
		TBody(
			g.MapIter(result.Hits, func(hit index.DocumentMatch) g.Node {
				if m, ok := hit.Data.(nix.Package); ok {
					return packageRow(hit, m)
				}

				return emptyPackageRow(hit)
			}),
		),
	)
}

func emptyPackageRow(hit index.DocumentMatch) g.Node {
	return Tr(
		Td(
			openDialogLink(hit.ID),
		),
		Td(
			g.Text(hit.ID),
		),
		Td(Class("description"),
			g.Text("No description available"),
		),
		g.If(config.DevMode,
			Td(Class("score"),
				Score(hit),
			),
		),
	)
}

func packageRow(hit index.DocumentMatch, p nix.Package) g.Node {
	return Tr(
		Td(
			openDialogLink(p.Attribute),
		),
		Td(
			g.Text(p.Name),
		),
		Td(Class("description"),
			g.Text(p.Description),
		),
		g.If(config.DevMode,
			Td(Class("score"),
				Score(hit),
			),
		),
	)
}
