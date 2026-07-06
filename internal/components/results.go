package components

import (
	"maps"
	"slices"
	"strings"

	"github.com/blevesearch/bleve/v2/search"

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

	qs := r.SearchNav.URL.RawQuery
	if r.Source.Importer != config.All {
		if qs != "" {
			qs += "&scoped"
		} else {
			qs = "scoped"
		}
	}
	var content g.Node
	switch r.Source.Importer {
	case config.Options:
		content = Options(r.Results, qs)
	case config.Packages:
		content = Packages(r.Results, qs)
	case config.All:
		content = Combined(r.Results, qs)
	}

	return g.Group([]g.Node{
		Facets(r),
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

func Facets(r ResultData) g.Node {
	return Nav(
		Class("facets"),
		Summary(
			H3(g.Text("Filters")),
		),
		g.Map(slices.Sorted(maps.Keys(r.Results.Facets)), func(name string) g.Node {
			facet := r.Results.Facets[name]

			return Div(Class("facet"),
				H4(g.Text(name)),
				Ul(
					g.Map(facet.Terms.Terms(), func(term *search.TermFacet) g.Node {
						t := term.Term
						if t == "" {
							t = "None"
						}

						return Term(
							g.If(
								r.SearchNav.HasFacet(facet.Field, term.Term),
								A(
									Href(r.SearchNav.RemoveFacet(facet.Field, term.Term)),
									TitleAttr("remove"),
									g.Text(breakable(t)),
								),
								A(
									Href(r.SearchNav.SetFacet(facet.Field, term.Term)),
									Rel("nofollow"),
									g.Text(breakable(t)),
								),
							),
							Span(
								Class("count"),
								g.Textf("%d", term.Count),
							),
						)
					}),
					g.If(facet.Missing > 0,
						Li(Class("term"),
							A(
								Href(r.SearchNav.SetFacet(facet.Field, "")),
								Rel("nofollow"),
								g.Text("None"),
							),
							Span(Class("count"), g.Textf("%d", facet.Missing)),
						),
					),
				),
			)
		}),
	)
}

const breakOpportunity = string(rune(0x200b))

func breakable(s string) string {
	return strings.Join(
		strings.Split(s, "."),
		"."+breakOpportunity,
	)
}

func Term(children ...g.Node) g.Node {
	return Li(Class("term"), g.Group(children))
}

func ResultsPage(r ResultData) g.Node {
	r.CanonicalURL = r.SearchNav.WithoutFacets()

	return SearchPage(r.TemplateData, r, Results(r))
}

func DetailResultsPage(r ResultData) g.Node {
	return SearchPage(r.TemplateData, r, Results(r))
}

func openDialogLink(attr string, qs string) g.Node {
	href := attr
	if qs != "" {
		href += "?" + qs
	}

	return A(Class("open-dialog"), Href(href), g.Text(attr))
}

func openCombinedDialogLink(attr string, qs string) g.Node {
	href := "/" + attr
	if qs != "" {
		href += "?" + qs
	}

	return A(Class("open-dialog"), Href(href), g.Text(attr))
}
