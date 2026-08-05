package components

import (
	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/nix"

	g "alin.ovh/gomponents"
	. "alin.ovh/gomponents/html"
)

func Options(result *index.Result, qs string) g.Node {
	return Table(
		THead(
			Tr(
				Th(Scope("col"), g.Text("Title")),
				Th(Scope("col"), Class("description"), g.Text("Description")),
				g.If(config.DevMode,
					Th(Scope("col"), Class("score"), g.Text("Score")),
				),
			),
		),
		TBody(
			g.MapIter(result.Hits, func(hit index.DocumentMatch) g.Node {
				if m, ok := hit.Data.(nix.Option); ok {
					return optionRow(hit, m, qs)
				}

				return emptyOptionRow(hit)
			}),
		),
	)
}

func emptyOptionRow(hit index.DocumentMatch) g.Node {
	return Tr(
		Td(
			g.Text(hit.ID),
		),
		Td(Class("description"),
			g.Text("No description available"),
			Dialog(ID(hit.ID), ClosedBy("any")),
		),
		g.If(config.DevMode,
			Td(Class("score"),
				Score(hit),
			),
		),
	)
}

func optionRow(hit index.DocumentMatch, o nix.Option, qs string) g.Node {
	return Tr(
		Td(
			openDialogLink(o.Name, qs),
		),
		Td(Class("description"),
			firstSentence(o.Description),
			Dialog(ID(o.Name), ClosedBy("any")),
		),
		g.If(config.DevMode,
			Td(Class("score"),
				Score(hit),
			),
		),
	)
}
