//nolint:wrapcheck
package filepath

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

// Node represents a file path reference node in markdown AST.
type Node struct {
	ast.BaseInline
	FilePath []byte
}

// Inline implements ast.Inline interface.
func (n *Node) Inline() bool {
	return true
}

// Kind implements ast.Node.Kind interface.
func (n *Node) Kind() ast.NodeKind {
	return KindFilePath
}

// Dump implements ast.Node.Dump interface.
func (n *Node) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"FilePath": string(n.FilePath),
	}, nil)
}

// KindFilePath is a NodeKind for file path nodes.
var KindFilePath = ast.NewNodeKind("FilePath")

// Parser is a Goldmark inline parser for parsing file path nodes.
//
// File path references have the format {file}`/etc/passwd` which will be rendered as
// a span with the file path.
type Parser struct{}

var _ parser.InlineParser = (*Parser)(nil)

// Trigger reports characters that trigger this parser.
func (*Parser) Trigger() []byte {
	return []byte{'{'}
}

// Parse parses a file path node.
func (p *Parser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, segment := block.PeekLine()

	// Check if we have enough characters for {file}`
	if len(line) < 7 || line[0] != '{' {
		return nil
	}

	// Check for {file}` prefix
	if !bytes.HasPrefix(line, []byte("{file}`")) {
		return nil
	}

	// Skip the {file}` prefix
	line = line[7:]
	start := segment.Start + 7

	// Find the closing backtick
	endPos := bytes.IndexByte(line, '`')
	if endPos < 0 {
		return nil
	}

	// Extract the file path
	filePath := line[:endPos]
	if len(filePath) == 0 {
		return nil
	}

	// Create the node
	n := &Node{
		FilePath: filePath,
	}

	// Set the text segment to include the entire {file}`path` text
	textSegment := text.NewSegment(segment.Start, start+endPos+1)
	n.AppendChild(n, ast.NewTextSegment(textSegment))

	// Advance the reader past this node
	block.Advance(7 + endPos + 1) // {file}` + path + `

	return n
}

// HTMLRenderer is a renderer for file path nodes.
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
	reg.Register(KindFilePath, r.renderFilePath)
}

func (r *HTMLRenderer) renderFilePath(
	w util.BufWriter,
	source []byte, //nolint:revive
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*Node)

	if _, err := w.WriteString(`<span class="file-path">`); err != nil {
		return ast.WalkStop, err
	}

	if _, err := w.Write(n.FilePath); err != nil {
		return ast.WalkStop, err
	}

	if _, err := w.WriteString("</span>"); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkSkipChildren, nil
}

// Extension is a goldmark extension for file path references.
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

// New returns a new file path extension.
func New() goldmark.Extender {
	return &Extension{}
}
