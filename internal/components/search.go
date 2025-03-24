package components

import (
	"time"

	g "go.alanpearce.eu/gomponents"
	. "go.alanpearce.eu/gomponents/html"
	"go.alanpearce.eu/searchix/internal/config"
)

func SearchForm(tdata TemplateData, r ResultData) g.Node {
	return Search(Form(
		ID("search"),
		FieldSet(
			Legend(
				ID("legend"),
				H2(g.Textf("%s search", sourceNameAndType(tdata.Source))),
			),
			Input(
				ID("query"),
				Aria("labelledby", "legend"),
				MinLength("2"),
				Name("query"),
				Type("search"),
				Value(r.Query),
				AutoFocus(),
				g.Attr("spellcheck", "false"),
				g.Attr("autocapitalize", "none"),
			),
			Button(g.Text("Search")),
		),
	))
}

func SearchPage(tdata TemplateData, r ResultData, children ...g.Node) g.Node {
	return Page(
		tdata,
		P(
			g.Text("Search Nix packages and options from "),
			MapCommaList(tdata.Sources, func(source *config.Source) g.Node {
				return A(Href(source.Repo.String()), g.Text(source.Name))
			}),
		),
		g.If(Indexing.InProgress,
			P(Class("notice"),
				g.Text("Indexing in progress, started "),
				Time(
					DateTime(Indexing.StartedAt.Format(time.RFC3339)),
					Title(Indexing.StartedAt.Format(time.DateTime)),
					g.Text(time.Since(Indexing.StartedAt).Round(time.Second).String()),
				),
				g.Text(" ago. "),
				g.If(!Indexing.FinishedAt.IsZero(),
					g.Group([]g.Node{
						g.Text("Last run took "),
						Time(
							DateTime(Indexing.FinishedAt.Format(time.RFC3339)),
							Title(Indexing.FinishedAt.Format(time.DateTime)),
							g.Text(time.Since(Indexing.FinishedAt).Round(time.Minute).String()),
						),
					}),
				),
			),
			P(
				g.Text("Indexing last ran "),
				Time(
					DateTime(Indexing.FinishedAt.Format(time.RFC3339)),
					Title(Indexing.FinishedAt.Format(time.DateTime)),
					g.Textf("%.0f hours ago", time.Since(Indexing.FinishedAt).Hours()),
				),
				g.Text(", will run again in "),
				Time(
					DateTime(Indexing.NextRun.Format(time.RFC3339)),
					Title(Indexing.NextRun.Format(time.DateTime)),
					g.Textf("%.0f hours", time.Until(Indexing.NextRun).Hours()),
				),
				g.Text("."),
			),
		),
		script(tdata.Assets.ByPath["/static/search.js"]),
		SearchForm(tdata, r),
		Section(
			ID("results"),
			Role("list"),
			Aria("label", "search results"),
			g.Group(children),
		),
		Dialog(
			ID("dialog"),
			Button(AutoFocus(), g.Text("Close")),
		),
		NoScript(
			P(
				Class("notice"),
				g.Text("Everything should work fine without JavaScript. If that is not the case, "),
				A(Href("https://todo.sr.ht/~alanpearce/searchix"), g.Text("report an issue")),
			),
		),
	)
}

func MapCommaList[T any](items []T, fn func(T) g.Node) g.Node {
	out := make([]g.Node, (len(items) * 2))
	j := 0
	last := len(items) - 2
	for i := range items {
		out[j] = fn(items[i])
		j++
		if i <= last {
			if i == last {
				out[j] = g.Text(" and ")
			} else {
				out[j] = g.Text(", ")
			}
		}
		j++
	}

	return g.Group(out)
}
