package components

import (
	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/nix"

	g "alin.ovh/gomponents"
	. "alin.ovh/gomponents/html"
)

func CombinedData(data nix.Importable) g.Node {
	switch data := data.(type) {
	case nix.Option:
		return firstSentence(data.Description)
	case nix.Package:
		return g.Text(firstSentence(data.Description))
	}

	return g.Text("No description Available")
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
						openCombinedDialogLink(hit.ID),
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
