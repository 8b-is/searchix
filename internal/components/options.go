package components

import (
	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/nix"

	g "alin.ovh/gomponents"
	. "alin.ovh/gomponents/html"
)

func Options(result *index.Result) g.Node {
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
				if m := convertMatch[nix.Option](hit.Data); m != nil {
					return optionRow(hit, *m)
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
			Dialog(ID(hit.ID)),
		),
		g.If(config.DevMode,
			Td(Class("score"),
				Score(hit),
			),
		),
	)
}

func optionRow(hit index.DocumentMatch, o nix.Option) g.Node {
	return Tr(
		Td(
			openDialogLink(o.Name),
		),
		Td(Class("description"),
			firstSentence(o.Description),
			Dialog(ID(o.Name)),
		),
		g.If(config.DevMode,
			Td(Class("score"),
				Score(hit),
			),
		),
	)
}
