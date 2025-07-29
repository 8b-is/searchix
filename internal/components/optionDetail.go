package components

import (
	"alin.ovh/searchix/internal/nix"

	g "alin.ovh/gomponents"
	. "alin.ovh/gomponents/html"
)

func OptionDetail(option nix.Option) g.Node {
	return g.Group([]g.Node{
		H2(g.Text(option.Name)),
		option.Description,
		Dl(
			g.If(option.Type != "",
				g.Group([]g.Node{
					Dt(g.Text("Type")),
					Dd(Code(g.Text(option.Type))),
				}),
			),
			g.Iff(option.Default != nil,
				func() g.Node {
					return g.Group([]g.Node{
						Dt(g.Text("Default")),
						Dd(
							g.If(option.Default.Markdown != "",
								option.Default.Markdown,
								Pre(Code(g.Text(option.Default.Plain))),
							),
						),
					})
				},
			),
			g.Iff(option.Example != nil,
				func() g.Node {
					return g.Group([]g.Node{
						Dt(g.Text("Example")),
						Dd(
							g.If(option.Example.Markdown != "",
								option.Example.Markdown,
								Pre(Code(g.Text(option.Example.Plain))),
							),
						),
					})
				},
			),
			g.If(option.RelatedPackages != "",
				g.Group([]g.Node{
					Dt(g.Text("Related Packages")),
					Dd(
						option.RelatedPackages,
					),
				}),
			),
			g.If(len(option.Declarations) > 0,
				g.Group([]g.Node{
					Dt(g.Text("Declared")),
					g.Map(option.Declarations, func(d nix.Link) g.Node {
						return Dd(
							A(Href(d.URL), g.Text(d.Name)),
						)
					}),
				}),
			),
		),
	})
}

func OptionDetailPage(tdata TemplateData, option nix.Option) g.Node {
	return Page(tdata, OptionDetail(option))
}
