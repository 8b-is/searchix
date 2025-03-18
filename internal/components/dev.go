package components

import (
	"strconv"

	"go.alanpearce.eu/searchix/internal/index"

	g "go.alanpearce.eu/gomponents"
	. "go.alanpearce.eu/gomponents/html"
)

func Score(h index.DocumentMatch) g.Node {
	return g.Group([]g.Node{
		A(
			Class("open-sibling-dialog"),
			g.Text(strconv.FormatFloat(h.Score, 'f', 2, 64)),
		),
		Dialog(
			Button(AutoFocus(), g.Text("Close")),
			Pre(g.Text(h.Expl.String())),
		),
	})
}
