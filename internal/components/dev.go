package components

import (
	"strconv"

	"alin.ovh/searchix/internal/index"

	g "alin.ovh/gomponents"
	. "alin.ovh/gomponents/html"
)

func Score(h index.DocumentMatch) g.Node {
	return g.Group([]g.Node{
		A(
			Class("open-sibling-dialog"),
			g.Text(strconv.FormatFloat(h.Score, 'f', 2, 64)),
		),
		Dialog(
			ClosedBy("any"),
			Button(AutoFocus(), g.Text("Close")),
			Pre(g.Text(h.Expl.String())),
		),
	})
}
