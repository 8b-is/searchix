package components

import (
	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/index"
	"go.alanpearce.eu/searchix/internal/nix"

	g "go.alanpearce.eu/gomponents"
	. "go.alanpearce.eu/gomponents/html"
)

func CombinedData(data nix.Importable) g.Node {
	switch data.(type) {
	case nix.Option:
		if o := convertMatch[nix.Option](data); o != nil {
			return firstSentence(o.Description)
		}
	case nix.Package:
		if p := convertMatch[nix.Package](data); p != nil {
			return g.Text(firstSentence(p.Description))
		}
	}

	return g.Text("")
}

func Combined(result *index.Result) g.Node {
	return Table(
		THead(
			Tr(
				Th(Scope("col"), g.Text("Attribute")),
				Th(Scope("col"), Class("description"), g.Text("Description")),
				g.If(config.DevMode,
					Th(Scope("col"), Class("score"), g.Text("Score")),
				),
			),
		),
		TBody(
			g.MapIter(result.Hits, func(hit index.DocumentMatch) g.Node {
				return Tr(
					Td(
						openCombinedDialogLink(nix.GetKey(hit.Data)),
					),
					Td(Class("description"),
						CombinedData(hit.Data),
					),
					g.If(config.DevMode,
						Td(Class("score"),
							Score(hit),
						),
					),
				)
			}),
		),
	)
}
