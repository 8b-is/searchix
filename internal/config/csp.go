package config

// Code generated  DO NOT EDIT.

import (
	"github.com/crewjam/csp"
)

type CSP struct {
	BaseURI                 []string                     `csp:"base-uri" toml:"base-uri"`
	BlockAllMixedContent    bool                         `csp:"block-all-mixed-content" toml:"block-all-mixed-content"`
	ChildSrc                []string                     `csp:"child-src" toml:"child-src"`
	ConnectSrc              []string                     `csp:"connect-src" toml:"connect-src"`
	DefaultSrc              []string                     `csp:"default-src" toml:"default-src"`
	FontSrc                 []string                     `csp:"font-src" toml:"font-src"`
	FormAction              []string                     `csp:"form-action" toml:"form-action"`
	FrameAncestors          []string                     `csp:"frame-ancestors" toml:"frame-ancestors"`
	FrameSrc                []string                     `csp:"frame-src" toml:"frame-src"`
	ImgSrc                  []string                     `csp:"img-src" toml:"img-src"`
	ManifestSrc             []string                     `csp:"manifest-src" toml:"manifest-src"`
	MediaSrc                []string                     `csp:"media-src" toml:"media-src"`
	NavigateTo              []string                     `csp:"navigate-to" toml:"navigate-to"`
	ObjectSrc               []string                     `csp:"object-src" toml:"object-src"`
	PluginTypes             []string                     `csp:"plugin-types" toml:"plugin-types"`
	PrefetchSrc             []string                     `csp:"prefetch-src" toml:"prefetch-src"`
	Referrer                csp.ReferrerPolicy           `csp:"referrer" toml:"referrer"`
	ReportTo                string                       `csp:"report-to" toml:"report-to"`
	ReportURI               string                       `csp:"report-uri" toml:"report-uri"`
	RequireSRIFor           []csp.RequireSRIFor          `csp:"require-sri-for" toml:"require-sri-for"`
	RequireTrustedTypesFor  []csp.RequireTrustedTypesFor `csp:"require-trusted-types-for" toml:"require-trusted-types-for"`
	Sandbox                 csp.Sandbox                  `csp:"sandbox" toml:"sandbox"`
	ScriptSrc               []string                     `csp:"script-src" toml:"script-src"`
	ScriptSrcAttr           []string                     `csp:"script-src-attr" toml:"script-src-attr"`
	ScriptSrcElem           []string                     `csp:"script-src-elem" toml:"script-src-elem"`
	StyleSrc                []string                     `csp:"style-src" toml:"style-src"`
	StyleSrcAttr            []string                     `csp:"style-src-attr" toml:"style-src-attr"`
	StyleSrcElem            []string                     `csp:"style-src-elem" toml:"style-src-elem"`
	TrustedTypes            []string                     `csp:"trusted-types" toml:"trusted-types"`
	UpgradeInsecureRequests bool                         `csp:"upgrade-insecure-requests" toml:"upgrade-insecure-requests"`
	WorkerSrc               []string                     `csp:"worker-src" toml:"worker-src"`
}

func (c *CSP) String() string {
	return csp.Header(*c).String()
}
