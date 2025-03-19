package components

import (
	"time"

	g "go.alanpearce.eu/gomponents"
	. "go.alanpearce.eu/gomponents/html"
)

func Search(tdata TemplateData, r ResultData) g.Node {
	return Form(
		ID("search"),
		Role("search"),
		FieldSet(
			Legend(
				ID("legend"),
				H2(g.Textf("%s search", sourceNameAndType(tdata.Source))),
			),
			Input(
				ID("query"),
				Aria("labelledby", "legend"),
				Name("query"),
				Type("search"),
				Value(r.Query),
				AutoFocus(),
				g.Attr("spellcheck", "false"),
				g.Attr("autocapitalize", "none"),
			),
			Button(g.Text("Search")),
		),
	)
}

func SearchPage(tdata TemplateData, r ResultData, children ...g.Node) g.Node {
	return Page(
		tdata,
		P(
			g.Text("Search Nix packages and options from "),
			A(Href("https://nixos.org"), g.Text("NixOS")),
			g.Text(", "),
			A(Href("https://github.com/LnL7/nix-darwin"), g.Text("nix-darwin")),
			g.Text(" and "),
			A(Href("https://github.com/nix-community/home-manager"), g.Text("home-manager")),
		),
		g.If(Indexing.InProgress,
			P(Class("notice"),
				g.Text("Indexing in progress, started "),
				Time(
					DateTime(Indexing.StartedAt.Format(time.RFC3339)),
					Title(Indexing.StartedAt.Format(time.RFC3339)),
					g.Text(time.Since(Indexing.StartedAt).Round(time.Second).String()),
				),
				g.Text(" ago. "),
				g.If(!Indexing.FinishedAt.IsZero(),
					g.Text("Last run took "),
					Time(
						DateTime(Indexing.FinishedAt.Format(time.RFC3339)),
						Title(Indexing.FinishedAt.Format(time.RFC3339)),
						g.Text(time.Since(Indexing.FinishedAt).Round(time.Minute).String()),
					),
				),
			),
			P(
				g.Text("Indexing last ran "),
				Time(
					DateTime(Indexing.FinishedAt.Format(time.RFC3339)),
					Title(Indexing.FinishedAt.Format(time.RFC3339)),
					g.Text(time.Since(Indexing.FinishedAt).Round(time.Minute).String()),
				),
				g.Text(" ago, will run again in "),
				Time(
					DateTime(Indexing.NextRun.Format(time.RFC3339)),
					Title(Indexing.NextRun.Format(time.RFC3339)),
					g.Text(time.Until(Indexing.NextRun).Round(time.Minute).String()),
				),
				g.Text("."),
			),
		),
		script(tdata.Assets.ByPath["/static/search.js"]),
		Search(tdata, r),
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
