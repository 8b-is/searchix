//nolint:wrapcheck
package manpage

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Node represents a manpage reference node in markdown AST.
type Node struct {
	ast.BaseInline
	ManPage []byte
}

// Inline implements ast.Inline interface.
func (n *Node) Inline() bool {
	return true
}

// Kind implements ast.Node.Kind interface.
func (n *Node) Kind() ast.NodeKind {
	return KindManPage
}

// Dump implements ast.Node.Dump interface.
func (n *Node) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"ManPage": string(n.ManPage),
	}, nil)
}

// KindManPage is a NodeKind for manpage nodes.
var KindManPage = ast.NewNodeKind("ManPage")

// manPagePattern matches a manpage reference with section number in parentheses
// e.g., "nix.conf(5)", "git(1)", etc.
var manPagePattern = regexp.MustCompile(`^([^()]+)\(([1-9])\)$`)

// Parser is a Goldmark inline parser for parsing manpage nodes.
//
// Manpage references have the format {manpage}`nix.conf(5)` which will be rendered as
// a link to the manpage with the appropriate section.
type Parser struct{}

var _ parser.InlineParser = (*Parser)(nil)

// Trigger reports characters that trigger this parser.
func (*Parser) Trigger() []byte {
	return []byte{'{'}
}

// Parse parses a manpage node.
func (p *Parser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, segment := block.PeekLine()

	// Check if we have enough characters for {manpage}`
	if len(line) < 11 || line[0] != '{' {
		return nil
	}

	// Check for {manpage}` prefix
	if !bytes.HasPrefix(line, []byte("{manpage}`")) {
		return nil
	}

	// Skip the {manpage}` prefix
	line = line[10:]
	start := segment.Start + 10

	// Find the closing backtick
	endPos := bytes.IndexByte(line, '`')
	if endPos < 0 {
		return nil
	}

	// Extract the manpage reference
	manPageRef := line[:endPos]
	if len(manPageRef) == 0 {
		return nil
	}

	// Create the node
	n := &Node{
		ManPage: manPageRef,
	}

	// Set the text segment to include the entire {manpage}`ref` text
	textSegment := text.NewSegment(segment.Start, start+endPos+1)
	n.AppendChild(n, ast.NewTextSegment(textSegment))

	// Advance the reader past this node
	block.Advance(10 + endPos + 1) // {manpage}` + ref + `

	return n
}

// HTMLRenderer is a renderer for manpage nodes.
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
	reg.Register(KindManPage, r.renderManPage)
}

func (r *HTMLRenderer) renderManPage(
	w util.BufWriter,
	source []byte, //nolint:revive
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*Node)
	manPageRef := string(n.ManPage)

	// Parse the manpage reference to extract name and section
	matches := manPagePattern.FindStringSubmatch(manPageRef)
	if len(matches) != 3 {
		// If it doesn't match the pattern, just render it as plain text
		if _, err := w.WriteString(manPageRef); err != nil {
			return ast.WalkStop, err
		}

		return ast.WalkSkipChildren, nil
	}

	name := matches[1]
	section := matches[2]

	// Generate the HTML link to the manpage
	href := fmt.Sprintf("/man/%s/%s", section, name)
	link := fmt.Sprintf(`<a class="manpage" href="%s">%s</a>`, href, manPageRef)

	if _, err := w.WriteString(link); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkSkipChildren, nil
}

// Extension is a goldmark extension for manpage references.
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

// New returns a new manpage extension.
func New() goldmark.Extender {
	return &Extension{}
}
