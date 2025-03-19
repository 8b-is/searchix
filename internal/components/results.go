package components

import (
	"go.alanpearce.eu/searchix/internal/config"

	g "go.alanpearce.eu/gomponents"
	. "go.alanpearce.eu/gomponents/html"
)

func Results(r ResultData) g.Node {
	if r.Query == "" {
		return Br()
	}

	if r.Results == nil || r.Results.Total == 0 {
		return Span(Role("status"), g.Text("Nothing found"))
	}

	var content g.Node
	if r.Source != nil {
		switch r.Source.Importer {
		case config.Options:
			content = Options(r.Results)
		case config.Packages:
			content = Packages(r.Results)
		}
	} else {
		content = Combined(r.Results)
	}

	return g.Group([]g.Node{
		content,
		Footer(
			g.Attr("aria-label", "pagination"),
			Nav(
				ID("pagination"),
				g.If(r.Prev != "",
					A(Class("button"), Href(r.Prev), Rel("prev"), g.Text("Prev")),
				),
				g.If(r.Next != "",
					A(Class("button"), Href(r.Next), Rel("next"), g.Text("Next")),
				),
			),
			Span(
				Role("status"),
				g.Textf("%d results", r.Results.Total),
			),
			g.Text(" "),
			g.If(r.Next != r.Prev && r.Results.Total < config.MaxResultsShowAll,
				A(Href(r.All), g.Text("Show All")),
			),
		),
	})
}

func ResultsPage(r ResultData) g.Node {
	return SearchPage(r.TemplateData, r, Results(r))
}

func openDialogLink(attr string) g.Node {
	return A(Class("open-dialog"), Href(attr), g.Text(attr))
}

func openCombinedDialogLink(attr string) g.Node {
	return A(Class("open-dialog"), Href("/"+attr), g.Text(attr))
}
