//nolint:wrapcheck
package variable

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

// Node represents a variable reference node in markdown AST.
type Node struct {
	ast.BaseInline
	VarName []byte
}

// Inline implements ast.Inline interface.
func (n *Node) Inline() bool {
	return true
}

// Kind implements ast.Node.Kind interface.
func (n *Node) Kind() ast.NodeKind {
	return KindVar
}

// Dump implements ast.Node.Dump interface.
func (n *Node) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"VarName": string(n.VarName),
	}, nil)
}

// KindVar is a NodeKind for variable nodes.
var KindVar = ast.NewNodeKind("Var")

// Parser is a Goldmark inline parser for parsing variable nodes.
//
// Variable references have the format {var}`pkgs.virtualbox` which will be rendered as
// a var HTML element with the variable name.
type Parser struct{}

var _ parser.InlineParser = (*Parser)(nil)

// Trigger reports characters that trigger this parser.
func (*Parser) Trigger() []byte {
	return []byte{'{'}
}

// Parse parses a variable node.
func (p *Parser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, segment := block.PeekLine()

	// Check if we have enough characters for {var}`
	if len(line) < 6 || line[0] != '{' {
		return nil
	}

	// Check for {var}` prefix
	if !bytes.HasPrefix(line, []byte("{var}`")) {
		return nil
	}

	// Skip the {var}` prefix
	line = line[6:]
	start := segment.Start + 6

	// Find the closing backtick
	endPos := bytes.IndexByte(line, '`')
	if endPos < 0 {
		return nil
	}

	// Extract the variable name
	varName := line[:endPos]
	if len(varName) == 0 {
		return nil
	}

	// Create the node
	n := &Node{
		VarName: varName,
	}

	// Set the text segment to include the entire {var}`name` text
	textSegment := text.NewSegment(segment.Start, start+endPos+1)
	n.AppendChild(n, ast.NewTextSegment(textSegment))

	// Advance the reader past this node
	block.Advance(6 + endPos + 1) // {var}` + name + `

	return n
}

// HTMLRenderer is a renderer for variable nodes.
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
	reg.Register(KindVar, r.renderVar)
}

func (r *HTMLRenderer) renderVar(
	w util.BufWriter,
	source []byte, //nolint:revive
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*Node)

	if _, err := w.WriteString("<var>"); err != nil {
		return ast.WalkStop, err
	}

	if _, err := w.Write(n.VarName); err != nil {
		return ast.WalkStop, err
	}

	if _, err := w.WriteString("</var>"); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkSkipChildren, nil
}

// Extension is a goldmark extension for variable references.
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

// New returns a new variable extension.
func New() goldmark.Extender {
	return &Extension{}
}
