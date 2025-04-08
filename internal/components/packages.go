package components

import (
	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/index"
	"go.alanpearce.eu/searchix/internal/nix"

	g "go.alanpearce.eu/gomponents"
	. "go.alanpearce.eu/gomponents/html"
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
				if m := convertMatch[nix.Package](hit.Data); m != nil {
					return packageRow(hit, *m)
				}

				return nil
			}),
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
