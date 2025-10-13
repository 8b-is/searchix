package opensearch

import (
	"encoding/xml"

	"alin.ovh/searchix/internal/config"
)

type Description struct {
	XMLName     xml.Name `xml:"http://a9.com/-/spec/opensearch/1.1/ OpenSearchDescription"`
	Description string
	LongName    string
	ShortName   string
	Image       Image
	SearchForm  *config.URL `xml:"http://www.mozilla.org/2006/browser/search/ SearchForm"`
	URL         URL         `xml:"Url"`
}

type URL struct {
	Method   string      `xml:"method,attr"`
	Template *config.URL `xml:"template,attr"`
	Type     string      `xml:"type,attr"`
}

type Image struct {
	Height  uint   `xml:"height,attr"`
	Width   uint   `xml:"width,attr"`
	Type    string `xml:"type,attr"`
	Content string
}
