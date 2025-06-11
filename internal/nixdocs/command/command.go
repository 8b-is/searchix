//nolint:wrapcheck
package command

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

// Node represents a command reference node in markdown AST.
type Node struct {
	ast.BaseInline
	Command []byte
}

// Inline implements ast.Inline interface.
func (n *Node) Inline() bool {
	return true
}

// Kind implements ast.Node.Kind interface.
func (n *Node) Kind() ast.NodeKind {
	return KindCommand
}

// Dump implements ast.Node.Dump interface.
func (n *Node) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Command": string(n.Command),
	}, nil)
}

// KindCommand is a NodeKind for command nodes.
var KindCommand = ast.NewNodeKind("Command")

// Parser is a Goldmark inline parser for parsing command nodes.
//
// Command references have the format {command}`name` which will be rendered as
// a span with the command name in code tags.
type Parser struct{}

var _ parser.InlineParser = (*Parser)(nil)

var openBrace = byte('{')

// Trigger reports characters that trigger this parser.
func (*Parser) Trigger() []byte {
	return []byte{openBrace}
}

// Parse parses a command node.
func (p *Parser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, segment := block.PeekLine()

	// Check if we have enough characters for {command}`
	if len(line) < 10 || line[0] != openBrace {
		return nil
	}

	// Check for {command}` prefix
	if !bytes.HasPrefix(line, []byte("{command}`")) {
		return nil
	}

	// Skip the {command}` prefix
	line = line[10:]
	start := segment.Start + 10

	// Find the closing backtick
	endPos := bytes.IndexByte(line, '`')
	if endPos < 0 {
		return nil
	}

	// Extract the command name
	commandName := line[:endPos]
	if len(commandName) == 0 {
		return nil
	}

	// Create the node
	n := &Node{
		Command: commandName,
	}

	// Set the text segment to include the entire {command}`name` text
	textSegment := text.NewSegment(segment.Start, start+endPos+1)
	n.AppendChild(n, ast.NewTextSegment(textSegment))

	// Advance the reader past this node
	block.Advance(10 + endPos + 1) // {command}` + name + `

	return n
}

// HTMLRenderer is a renderer for command nodes.
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
	reg.Register(KindCommand, r.renderCommand)
}

func (r *HTMLRenderer) renderCommand(
	w util.BufWriter,
	source []byte, //nolint:revive
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*Node)

	if _, err := w.WriteString(`<span class="command"><code>`); err != nil {
		return ast.WalkStop, err
	}

	if _, err := w.Write(n.Command); err != nil {
		return ast.WalkStop, err
	}

	if _, err := w.WriteString("</code></span>"); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkSkipChildren, nil
}

// Extension is a goldmark extension for command references.
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

// New returns a new command extension.
func New() goldmark.Extender {
	return &Extension{}
}
