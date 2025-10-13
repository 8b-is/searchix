package components

import (
	"net/url"

	"alin.ovh/searchix/frontend"
	"alin.ovh/searchix/internal/config"

	g "alin.ovh/gomponents"
	c "alin.ovh/gomponents/components"
	. "alin.ovh/gomponents/html"
)

func Page(tdata TemplateData, children ...g.Node) g.Node {
	return Doctype(
		HTML(
			Lang("en-GB"),
			Head(
				Meta(Charset("utf-8")),
				Meta(Name("viewport"), Content("width=device-width, initial-scale=1")),
				TitleEl(g.Text("Searchix"), g.If(config.DevMode, g.Text(" (Dev)"))),
				g.Map(tdata.Assets.Stylesheets, css),
				g.Raw(tdata.ExtraHeadHTML),
				Link(Rel("icon"), Href("/static/favicon.ico")),
				g.If(len(tdata.Sources) > 0,
					Link(
						Rel("search"),
						Type("application/opensearchdescription+xml"),
						TitleAttr("Searchix "+sourceNameAndType(nil)),
						Href(joinPath("opensearch.xml")),
					),
				),
				g.Map(tdata.Sources, func(source *config.Source) g.Node {
					return Link(
						Rel("search"),
						Type("application/opensearchdescription+xml"),
						TitleAttr("Searchix "+sourceNameAndType(source)),
						Href(joinPath("/", source.Importer.String(), source.Key, "opensearch.xml")),
					)
				}),
			),
			Body(
				Header(
					Nav(
						H1(A(Href("/"), g.Text("Searchix"))),
						g.If(len(tdata.Sources) > 0,
							A(
								c.Classes{
									"current": tdata.Source == nil,
								},
								g.If(
									tdata.Source == nil,
									Href("/"),
									Href(joinPathQuery("/", tdata.Query)),
								),
								g.Text("All"),
							),
						),
						g.Map(tdata.Sources, func(source *config.Source) g.Node {
							if tdata.Source != nil && tdata.Source.Name == source.Name {
								return A(
									Class("current"),
									Href(
										joinPath(
											"/",
											source.Importer.String(),
											source.Key,
											"search",
										),
									),
									g.Text(source.Name),
								)
							}

							return A(
								Href(
									joinPathQuery(
										joinPath(
											"/",
											source.Importer.String(),
											source.Key,
											"search",
										),
										tdata.Query,
									),
								),
								g.Text(source.Name),
							)
						}),
					),
				),
				Main(children...),
				Footer(
					P(
						g.If(config.Version != "",
							g.Group([]g.Node{
								g.Text("Searchix "),
								A(
									Href(
										"https://codeberg.org/alanpearce/searchix/src/tag/v"+config.Version+"/CHANGELOG.md",
									),
									TitleAttr("View changelog"),
									g.Textf("v%s", config.Version),
								),
								g.Text(" "),
							}),
						),
					),
					P(
						A(Href("https://codeberg.org/alanpearce/searchix"), g.Text("Source code")),
						g.Text(" "),
						A(Href("https://status.searchix.ovh"), g.Text("Status page")),
						g.Text(" "),
						A(
							Href("https://codeberg.org/alanpearce/searchix/issues"),
							g.Text("Report issues"),
						),
					),
					P(
						g.Text("Made by "),
						A(Href("https://alanpearce.eu"), g.Text("Alan Pearce")),
					),
				),
			),
		),
	)
}

func css(css *frontend.Asset) g.Node {
	return Link(Href(css.URL), Rel("stylesheet"))
}

func script(s *frontend.Asset) g.Node {
	return Script(Src(s.URL), Defer())
}

func sourceNameAndType(source *config.Source) string {
	if source == nil {
		return "Combined"
	}

	switch source.Importer {
	case config.Options:
		return source.Name + " " + source.Importer.String()
	case config.Packages:
		return source.Name
	}

	return ""
}

func joinPath(base string, parts ...string) string {
	u, err := url.JoinPath(base, parts...)
	if err != nil {
		panic(err)
	}

	return u
}

func joinPathQuery(path string, query string) string {
	if query == "" {
		return path
	}

	return path + "?query=" + url.QueryEscape(query)
}
