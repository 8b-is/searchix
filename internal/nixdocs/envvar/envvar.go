//nolint:wrapcheck
package envvar

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Node represents an environment variable reference node in markdown AST.
type Node struct {
	ast.BaseInline
	EnvVar []byte
}

// Inline implements ast.Inline interface.
func (n *Node) Inline() bool {
	return true
}

// Kind implements ast.Node.Kind interface.
func (n *Node) Kind() ast.NodeKind {
	return KindEnv
}

// Dump implements ast.Node.Dump interface.
func (n *Node) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"EnvVar": string(n.EnvVar),
	}, nil)
}

// KindEnv is a NodeKind for environment variable nodes.
var KindEnv = ast.NewNodeKind("Env")

// Parser is a Goldmark inline parser for parsing environment variable nodes.
//
// Environment variable references have the format {env}`XDG_DATA_DIRS` which will be rendered as
// a span with the environment variable name.
type Parser struct{}

var _ parser.InlineParser = (*Parser)(nil)

// Trigger reports characters that trigger this parser.
func (*Parser) Trigger() []byte {
	return []byte{'{'}
}

// Parse parses an environment variable node.
func (p *Parser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, segment := block.PeekLine()

	// Check if we have enough characters for {env}`
	if len(line) < 6 || line[0] != '{' {
		return nil
	}

	// Check for {env}` prefix
	if !bytes.HasPrefix(line, []byte("{env}`")) {
		return nil
	}

	// Skip the {env}` prefix
	line = line[6:]
	start := segment.Start + 6

	// Find the closing backtick
	endPos := bytes.IndexByte(line, '`')
	if endPos < 0 {
		return nil
	}

	// Extract the environment variable name
	envVarName := line[:endPos]
	if len(envVarName) == 0 {
		return nil
	}

	// Create the node
	n := &Node{
		EnvVar: envVarName,
	}

	// Set the text segment to include the entire {env}`name` text
	textSegment := text.NewSegment(segment.Start, start+endPos+1)
	n.AppendChild(n, ast.NewTextSegment(textSegment))

	// Advance the reader past this node
	block.Advance(6 + endPos + 1) // {env}` + name + `

	return n
}

// HTMLRenderer is a renderer for environment variable nodes.
type HTMLRenderer struct {
	html.Config
}

// NewHTMLRenderer returns a new HTMLRenderer.
func NewHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &HTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}

	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *HTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindEnv, r.renderEnv)
}

func (r *HTMLRenderer) renderEnv(
	w util.BufWriter,
	source []byte, //nolint:revive
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*Node)

	if _, err := w.WriteString(`<span class="envvar">`); err != nil {
		return ast.WalkStop, err
	}

	if _, err := w.Write(n.EnvVar); err != nil {
		return ast.WalkStop, err
	}

	if _, err := w.WriteString("</span>"); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkSkipChildren, nil
}

// Extension is a goldmark extension for environment variable references.
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

// New returns a new environment variable extension.
func New() goldmark.Extender {
	return &Extension{}
}
