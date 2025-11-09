package components

import (
	"alin.ovh/searchix/internal/nix"

	g "alin.ovh/gomponents"
	. "alin.ovh/gomponents/html"
)

func licenseName(l nix.License) string {
	if l.FullName != "" {
		return l.FullName
	}

	return l.Name
}

func PackageDetail(pkg nix.Package) g.Node {
	return g.Group([]g.Node{
		H2(
			g.If(pkg.Broken,
				Del(g.Text(pkg.Attribute)),
				g.Text(pkg.Attribute),
			),
		),
		g.If(pkg.LongDescription != "",
			pkg.LongDescription,
			P(g.Text(pkg.Description)),
		),
		Dl(
			g.If(pkg.MainProgram != "",
				g.Group([]g.Node{
					Dt(g.Text("Main Program")),
					Dd(Code(g.Text(pkg.MainProgram))),
				}),
			),
			g.If(len(pkg.Programs) > 0,
				g.Group([]g.Node{
					Dt(g.Text("Programs")),
					Dd(
						Ul(
							g.Map(pkg.Programs, func(p string) g.Node {
								return Li(Code(g.Text(p)))
							}),
						),
					),
				}),
			),
			g.If(len(pkg.Homepages) > 0,
				g.Group([]g.Node{
					Dt(g.Text("Homepage")),
					Dd(
						Ul(
							g.Map(pkg.Homepages, func(u string) g.Node {
								return Li(A(Href(u), g.Text(u)))
							}),
						),
					),
				}),
			),
			g.If(pkg.Version != "",
				g.Group([]g.Node{
					Dt(g.Text("Version")),
					Dd(g.Text(pkg.Version)),
				}),
			),
			g.If(len(pkg.Licenses) > 0,
				g.Group([]g.Node{
					Dt(g.Text("License")),
					Dd(
						Ul(
							g.Map(pkg.Licenses, func(l nix.License) g.Node {
								return Li(
									g.If(l.URL != "",
										A(Href(l.URL), g.Text(licenseName(l))),
										g.Text(licenseName(l)),
									),
									g.If(l.AppendixURL != "",
										A(Href(l.AppendixURL), g.Text("Appendix")),
									),
								)
							}),
						),
					),
				}),
			),
			g.If(len(pkg.Maintainers) > 0,
				g.Group([]g.Node{
					Dt(g.Text("Maintainers")),
					Dd(
						Ul(
							g.Map(pkg.Maintainers, func(m nix.Maintainer) g.Node {
								return Li(
									g.If(
										m.Github != "",
										A(
											Href(joinPath("https://github.com", m.Github)),
											g.Text(m.Name),
										),
										g.Text(m.Name),
									),
								)
							}),
						),
					),
				}),
			),
			g.If(len(pkg.Platforms) > 0,
				g.Group([]g.Node{
					Dt(g.Text("Platforms")),
					Dd(
						Ul(
							g.Map(pkg.Platforms, func(plat string) g.Node {
								return Li(g.Text(plat))
							}),
						),
					),
				}),
			),
			g.If(pkg.Definition != "",
				g.Group([]g.Node{
					Dt(g.Text("Defined")),
					Dd(A(Href(pkg.Definition), g.Text("Source"))),
				}),
			),
		),
	})
}
