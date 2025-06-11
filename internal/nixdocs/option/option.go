//nolint:wrapcheck
package option

import (
	"bytes"
	"html"
	"net/url"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	ghtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Node represents an option reference node in markdown AST.
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
	return KindOption
}

// Dump implements ast.Node.Dump interface.
func (n *Node) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Option": string(n.Option),
	}, nil)
}

// KindOption is a NodeKind for option nodes.
var KindOption = ast.NewNodeKind("Option")

// Parser is a Goldmark inline parser for parsing option nodes.
//
// Option references have the format {option}`name` which will be rendered as
// a link to the option with the name in code tags.
type Parser struct{}

var _ parser.InlineParser = (*Parser)(nil)

var openBrace = byte('{')

// Trigger reports characters that trigger this parser.
func (*Parser) Trigger() []byte {
	return []byte{openBrace}
}

// Parse parses an option node.
func (p *Parser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, segment := block.PeekLine()

	// Check if we have enough characters for {option}`
	if len(line) < 9 || line[0] != openBrace {
		return nil
	}

	// Check for {option}` prefix
	if !bytes.HasPrefix(line, []byte("{option}`")) {
		return nil
	}

	// Skip the {option}` prefix
	line = line[9:]
	start := segment.Start + 9

	// Find the closing backtick
	endPos := bytes.IndexByte(line, '`')
	if endPos < 0 {
		return nil
	}

	// Extract the option name
	optionName := line[:endPos]
	if len(optionName) == 0 {
		return nil
	}

	// Create the node
	n := &Node{
		Option: optionName,
	}

	// Set the text segment to include the entire {option}`name` text
	textSegment := text.NewSegment(segment.Start, start+endPos+1)
	n.AppendChild(n, ast.NewTextSegment(textSegment))

	// Advance the reader past this node
	block.Advance(9 + endPos + 1) // {option}` + name + `

	return n
}

// HTMLRenderer is a renderer for option nodes.
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
	reg.Register(KindOption, r.renderOption)
}

func (r *HTMLRenderer) renderOption(
	w util.BufWriter,
	source []byte, //nolint:revive
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*Node)

	if _, err := w.WriteString(
		`<a class="option" href="/?query=` + url.QueryEscape(string(n.Option)) + `"><code>`,
	); err != nil {
		return ast.WalkStop, err
	}

	if _, err := w.WriteString(html.EscapeString(string(n.Option))); err != nil {
		return ast.WalkStop, err
	}

	if _, err := w.WriteString("</code></a>"); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkSkipChildren, nil
}

// Extension is a goldmark extension for option references.
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

// New returns a new option extension.
func New() goldmark.Extender {
	return &Extension{}
}
