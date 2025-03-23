package components

import (
	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/index"
	"go.alanpearce.eu/searchix/internal/nix"

	g "go.alanpearce.eu/gomponents"
	. "go.alanpearce.eu/gomponents/html"
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

				return nil
			}),
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
			Td(
				Score(hit),
			),
		),
	)
}
