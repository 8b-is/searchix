package components

import (
	"alin.ovh/searchix/internal/config"

	g "alin.ovh/gomponents"
	. "alin.ovh/gomponents/html"
)

func Results(r ResultData) g.Node {
	if r.Results == nil {
		return Br()
	}

	if r.Results.Total == 0 {
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
				g.If(r.SearchNav.Prev != "",
					A(Class("button"), Href(r.SearchNav.Prev), Rel("prev"), g.Text("Prev")),
				),
				g.If(r.SearchNav.Next != "",
					A(Class("button"), Href(r.SearchNav.Next), Rel("next"), g.Text("Next")),
				),
			),
			Span(
				Role("status"),
				g.Textf("%d results", r.Results.Total),
			),
			g.Text(" "),
			g.If(r.SearchNav.All != "",
				A(Href(r.SearchNav.All), g.Text("Show All")),
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
