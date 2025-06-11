package manpage

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark"
)

func TestManPageExtension(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			New(),
		),
	)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic manpage reference",
			input:    "{manpage}`nix.conf(5)`",
			expected: "<p><a class=\"manpage\" href=\"/man/5/nix.conf\">nix.conf(5)</a></p>\n",
		},
		{
			name:     "manpage reference in sentence",
			input:    "See {manpage}`git(1)` for more information.",
			expected: "<p>See <a class=\"manpage\" href=\"/man/1/git\">git(1)</a> for more information.</p>\n",
		},
		{
			name:  "multiple manpage references",
			input: "Both {manpage}`ssh(1)` and {manpage}`sshd(8)` have configuration options.",
			expected: "<p>Both <a class=\"manpage\" href=\"/man/1/ssh\">ssh(1)</a> and " +
				"<a class=\"manpage\" href=\"/man/8/sshd\">sshd(8)</a> have configuration options.</p>\n",
		},
		{
			name:     "incomplete manpage reference - no closing backtick",
			input:    "{manpage}`missing.backtick",
			expected: "<p>{manpage}`missing.backtick</p>\n",
		},
		{
			name:     "incomplete manpage reference - empty reference",
			input:    "{manpage}``",
			expected: "<p>{manpage}``</p>\n",
		},
		{
			name:     "invalid manpage reference - no section",
			input:    "{manpage}`nixos`",
			expected: "<p>nixos</p>\n",
		},
		{
			name:     "invalid manpage reference - invalid section format",
			input:    "{manpage}`bash(x)`",
			expected: "<p>bash(x)</p>\n",
		},
		{
			name:  "manpage reference with code block",
			input: "Configure {manpage}`httpd.conf(5)` in your `apache` directory.",
			expected: "<p>Configure <a class=\"manpage\" href=\"/man/5/httpd.conf\">httpd.conf(5)</a> " +
				"in your <code>apache</code> directory.</p>\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(test.input), &buf); err != nil {
				t.Fatalf("Failed to convert markdown: %v", err)
			}

			if got := buf.String(); got != test.expected {
				t.Errorf("Expected:\n%q\nGot:\n%q", test.expected, got)
			}
		})
	}
}

func TestManPageNodeKind(t *testing.T) {
	n := &Node{}
	if kind := n.Kind(); kind != KindManPage {
		t.Errorf("Expected node kind %v, got %v", KindManPage, kind)
	}
	if !n.Inline() {
		t.Error("Expected node to be inline")
	}
}
