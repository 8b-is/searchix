package components

import (
	g "alin.ovh/gomponents"
	. "alin.ovh/gomponents/html"
)

func Error(tdata TemplateData) g.Node {
	return P(
		Class("notice error"),
		g.Textf("%d %s", tdata.Code, tdata.Message),
	)
}

func ErrorPage(tdata TemplateData) g.Node {
	return Page(tdata, Error(tdata))
}

func ErrorTemplate(tdata TemplateData) g.Node {
	return Page(tdata, P(
		Class("notice error"),
		g.Raw(tdata.Message),
	))
}
