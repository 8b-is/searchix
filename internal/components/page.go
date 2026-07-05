package components

import (
	"net/url"
	"strings"

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
				g.If(tdata.Description != "",
					g.Group([]g.Node{
						Meta(Name("description"), Content(tdata.Description)),
						Meta(g.Attr("property", "og:description"), Content(tdata.Description)),
					}),
				),
				Meta(g.Attr("property", "og:type"), Content("website")),
				Meta(g.Attr("property", "og:site_name"), Content("Searchix")),
				Meta(g.Attr("property", "og:title"), Content(ogTitle(tdata.Title))),
				g.Iff(tdata.BaseURL.URL != nil, func() g.Node {
					return g.Group([]g.Node{
						Meta(
							g.Attr("property", "og:image"),
							Content(tdata.BaseURL.JoinPath("apple-touch-icon.png").String()),
						),
					})
				}),
				TitleEl(
					g.Raw(tdata.Title),
					g.Text("Searchix"), g.If(config.DevMode, g.Text(" (Dev)")),
				),
				g.Map(tdata.Assets.Stylesheets, css),
				g.Raw(tdata.ExtraHeadHTML),
				g.If(len(tdata.Sources) > 0,
					Link(
						Rel("search"),
						Type("application/opensearchdescription+xml"),
						TitleAttr("Searchix "+tdata.Source.String()),
						Href(joinPath("opensearch.xml")),
					),
				),
				g.Map(tdata.Sources, func(source config.Source) g.Node {
					return Link(
						Rel("search"),
						Type("application/opensearchdescription+xml"),
						TitleAttr("Searchix "+source.String()),
						Href(joinPath("/", source.Importer.String(), source.Key, "opensearch.xml")),
					)
				}),
				g.Iff(tdata.CanonicalURL != "", func() g.Node {
					return Link(
						Rel("canonical"),
						Href(tdata.CanonicalURL),
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
									"current": tdata.Source.Importer == config.All,
								},
								g.If(
									tdata.Source.Importer == config.All,
									Href("/"),
									Href(joinPathQuery("/", tdata.Query)),
								),
								g.Text("All"),
							),
						),
						g.Map(tdata.Sources, func(source config.Source) g.Node {
							if tdata.Source.Importer != config.All && tdata.Source.Name == source.Name {
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
										"https://codeberg.org/alinnow/searchix/src/tag/v"+config.Version+"/CHANGELOG.md",
									),
									TitleAttr("View changelog"),
									g.Textf("v%s", config.Version),
								),
								g.Text(" "),
							}),
						),
					),
					P(
						A(Href("https://codeberg.org/alinnow/searchix"), g.Text("Source code")),
						g.Text(" "),
						A(Href("https://status.searchix.ovh"), g.Text("Status page")),
						g.Text(" "),
						A(
							Href("https://codeberg.org/alinnow/searchix/issues"),
							g.Text("Report issues"),
						),
					),
					P(
						g.Text("Made by "),
						A(Href("https://alanpearce.eu"), g.Text("Alin (Alan Pearce)")),
					),
				),
			),
		),
	)
}

func ogTitle(title string) string {
	t := strings.ReplaceAll(title, " &hyphen; ", " - ")
	t += "Searchix"

	return t
}

func css(css *frontend.Asset) g.Node {
	return Link(Href(css.ImmutablePath), Rel("stylesheet"))
}

func script(s *frontend.Asset) g.Node {
	return Script(Src(s.ImmutablePath), Defer())
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
