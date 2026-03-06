package components

import (
	"fmt"
	"strings"

	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/nix"

	g "alin.ovh/gomponents"
)

func Detail(thing index.Indexable) g.Node {
	switch t := thing.(type) {
	case nix.Option:
		return OptionDetail(t)
	case nix.Package:
		return PackageDetail(t)
	default:
		return nil
	}
}

const MaxDescriptionLength = 160

func TruncateWithEllipsis(text string, maxLength int) string {
	if len(text) > maxLength {
		if maxLength > 1 {
			return text[:maxLength-1] + "…"
		}

		return text
	}

	return text
}

func MetaDescription(thing index.Indexable) string {
	switch t := thing.(type) {
	case nix.Package:
		desc := t.Attribute
		if t.Version != "" {
			desc += " " + t.Version
		}
		if t.Description != "" {
			desc += " - " + t.Description
		}

		desc = TruncateWithEllipsis(desc, MaxDescriptionLength)

		return desc
	case nix.Option:
		desc := t.Name
		if t.Type != "" {
			desc = fmt.Sprintf("%s - %s option", desc, t.Type)
		}
		plain := strings.TrimSpace(string(t.Description))
		if plain != "" {
			desc += ". " + plain
		}

		desc = TruncateWithEllipsis(desc, MaxDescriptionLength)

		return desc
	default:
		return ""
	}
}

func DetailPage(tdata TemplateData, thing index.Indexable) g.Node {
	return Page(tdata, Detail(thing))
}
