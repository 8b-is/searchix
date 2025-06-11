//nolint:wrapcheck
package optlink

import (
	"html"
	"net/url"
	"regexp"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	ghtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Node represents an option link node in markdown AST.
type Node struct {
	ast.BaseInline
	Option []byte
}

// Inline implements ast.Inline interface.
func (n *Node) Inline() bool {
	return true
}

// Kind implements ast.Node.Kind interface.
func (n *Node) Kind() ast.NodeKind {
	return KindOptLink
}

// Dump implements ast.Node.Dump interface.
func (n *Node) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Option": string(n.Option),
	}, nil)
}

// KindOptLink is a NodeKind for option link nodes.
var KindOptLink = ast.NewNodeKind("OptLink")

// Parser is a Goldmark inline parser for parsing option link nodes.
//
// Option links have the format [](#opt-option.name) which will be rendered as
// a link to the option with the name.
type Parser struct{}

var _ parser.InlineParser = (*Parser)(nil)

var openBracket = byte('[')

// Trigger reports characters that trigger this parser.
func (*Parser) Trigger() []byte {
	return []byte{openBracket}
}

// optLinkPattern matches [](#opt-option.name)
var optLinkPattern = regexp.MustCompile(`^\[\]\(#opt-([^)]+)\)`)

// Parse parses an option link node.
func (p *Parser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, segment := block.PeekLine()

	// Check if we have enough characters for [](#opt-
	if len(line) < 8 || line[0] != openBracket {
		return nil
	}

	// Use regex to match the pattern [](#opt-option.name)
	matches := optLinkPattern.FindSubmatch(line)
	if matches == nil {
		return nil
	}

	// Extract the option name from the match
	optionName := matches[1]
	if len(optionName) == 0 {
		return nil
	}

	// Create the node
	n := &Node{
		Option: optionName,
	}

	// Set the text segment to include the entire [](#opt-option.name) text
	matchLength := len(matches[0])
	textSegment := text.NewSegment(segment.Start, segment.Start+matchLength)
	n.AppendChild(n, ast.NewTextSegment(textSegment))

	// Advance the reader past this node
	block.Advance(matchLength)

	return n
}

// HTMLRenderer is a renderer for option link nodes.
type HTMLRenderer struct {
	ghtml.Config
}

// NewHTMLRenderer returns a new HTMLRenderer.
func NewHTMLRenderer(opts ...ghtml.Option) renderer.NodeRenderer {
	r := &HTMLRenderer{
		Config: ghtml.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}

	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *HTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindOptLink, r.renderOptLink)
}

func (r *HTMLRenderer) renderOptLink(
	w util.BufWriter,
	source []byte, //nolint:revive
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*Node)
	optionName := string(n.Option)

	if _, err := w.WriteString(
		`<a class="option" href="/?query=` + url.QueryEscape(optionName) + `">`,
	); err != nil {
		return ast.WalkStop, err
	}

	if _, err := w.WriteString(html.EscapeString(optionName)); err != nil {
		return ast.WalkStop, err
	}

	if _, err := w.WriteString("</a>"); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkSkipChildren, nil
}

// Extension is a goldmark extension for option links.
type Extension struct{}

// Extend implements goldmark.Extender.Extend.
func (e *Extension) Extend(m goldmark.Markdown) {
	// Register the parser
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(new(Parser), 100),
		),
	)

	// Register the renderer
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewHTMLRenderer(), 500),
		),
	)
}

// New returns a new option link extension.
func New() goldmark.Extender {
	return &Extension{}
}
