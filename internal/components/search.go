package components

import (
	"fmt"
	"time"

	g "alin.ovh/gomponents"
	. "alin.ovh/gomponents/html"
	"github.com/dustin/go-humanize"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/importer"
)

func SearchForm(tdata TemplateData, r ResultData, children ...g.Node) g.Node {
	return Form(
		ID("search"),
		Search(
			FieldSet(
				Legend(
					ID("legend"),
					H2(g.Textf("%s search", tdata.Source.String())),
					A(
						Class("help"),
						Target("_blank"),
						Href("https://blevesearch.com/docs/Query-String-Query/"),
						g.Text("advanced query syntax"),
					),
				),
				Input(
					ID("query"),
					Aria("labelledby", "legend"),
					MinLength("2"),
					Name("query"),
					Type("search"),
					Value(r.Query),
					g.If(tdata.DetailContent == nil, AutoFocus()),
					g.Attr("spellcheck", "false"),
					g.Attr("autocapitalize", "none"),
				),
				Button(g.Text("Search")),
			),
		),
		g.Group(children),
	)
}

func SearchPage(tdata TemplateData, r ResultData, children ...g.Node) g.Node {
	return Page(
		tdata,
		P(
			g.Text("Search Nix packages and options from "),
			MapCommaList(tdata.Sources, func(source config.Source) g.Node {
				return A(Href(source.Repo.String()), g.Text(source.Name))
			}),
		),
		g.If(!importer.Job.StartedAt.IsZero(),
			P(Class("notice"),
				g.Text("Indexing in progress, started "),
				Time(
					DateTime(importer.Job.StartedAt.Format(time.RFC3339)),
					Title(importer.Job.StartedAt.Format(time.DateTime)),
					g.Text(time.Since(importer.Job.StartedAt).Round(time.Second).String()),
				),
				g.Text(" ago. "),
				g.Iff(!importer.Job.LastRun.FinishedAt.IsZero(), func() g.Node {
					duration := importer.Job.LastRun.FinishedAt.Sub(importer.Job.LastRun.StartedAt)

					return g.Group([]g.Node{
						g.Text("Last run took "),
						Time(
							DateTime(DurationDateTimeString(duration)),
							Title(duration.Round(time.Second).String()),
							g.Text(duration.Round(time.Minute).String()),
						),
					})
				}),
			),
			g.If(!importer.Job.LastRun.FinishedAt.IsZero(),
				P(
					g.Text("Indexing last ran "),
					Time(
						DateTime(importer.Job.LastRun.FinishedAt.Format(time.RFC3339)),
						Title(importer.Job.LastRun.FinishedAt.Format(time.DateTime)),
						g.Text(
							humanize.RelTime(
								importer.Job.LastRun.FinishedAt,
								time.Now(),
								"ago",
								"in the future",
							),
						),
					),
					g.Text(", will run again in "),
					Time(
						DateTime(importer.Job.NextRun.Format(time.RFC3339)),
						Title(importer.Job.NextRun.Format(time.DateTime)),
						g.Textf("%.0f hours", time.Until(importer.Job.NextRun).Hours()),
					),
					g.Text("."),
				),
				P(Class("notice error"), g.Text("Index missing")),
			),
		),
		script(tdata.Assets.ByPath["/search.js"]),
		SearchForm(tdata, r,
			Div(
				ID("results"),
				Role("list"),
				Aria("label", "search results"),
				g.Group(children),
			),
		),
		Dialog(
			ID("dialog"),
			ClosedBy("any"),
			g.If(tdata.DetailContent != nil, g.Attr("open")),
			g.If(tdata.SearchTitle != "", g.Attr("data-search-title", tdata.SearchTitle)),
			g.If(tdata.DetailContent != nil,
				A(Class("button"), Href(tdata.SearchURL), AutoFocus(), g.Text("Close")),
				Button(AutoFocus(), g.Text("Close")),
			),
			g.If(tdata.DetailContent != nil, tdata.DetailContent),
		),
		NoScript(
			P(
				Class("notice"),
				g.Text("Everything should work fine without JavaScript. If that is not the case, "),
				A(
					Href("https://codeberg.org/alinnow/searchix/issues"),
					g.Text("report an issue"),
				),
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

func DurationDateTimeString(d time.Duration) string {
	// PThHmMsS
	return fmt.Sprintf("PT%dH%dM%dS", d/time.Hour, d/time.Minute%60, d/time.Second%60)
}
