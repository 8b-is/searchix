package components

import (
	g "go.alanpearce.eu/gomponents"
	. "go.alanpearce.eu/gomponents/html"
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
