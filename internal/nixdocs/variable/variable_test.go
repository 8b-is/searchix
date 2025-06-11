package variable

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark"
)

func TestVarExtension(t *testing.T) {
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
			name:     "basic variable reference",
			input:    "{var}`pkgs.virtualbox`",
			expected: "<p><var>pkgs.virtualbox</var></p>\n",
		},
		{
			name:     "variable reference in sentence",
			input:    "Use {var}`nixpkgs.lib.attrsets` to manipulate attribute sets.",
			expected: "<p>Use <var>nixpkgs.lib.attrsets</var> to manipulate attribute sets.</p>\n",
		},
		{
			name:  "multiple variable references",
			input: "Both {var}`config.services.nginx` and {var}`config.services.postgresql` define service configurations.",
			expected: "<p>Both <var>config.services.nginx</var> and <var>" +
				"config.services.postgresql</var> define service configurations.</p>\n",
		},
		{
			name:     "incomplete variable reference - no closing backtick",
			input:    "{var}`MISSING_BACKTICK",
			expected: "<p>{var}`MISSING_BACKTICK</p>\n",
		},
		{
			name:     "incomplete variable reference - empty variable name",
			input:    "{var}``",
			expected: "<p>{var}``</p>\n",
		},
		{
			name:  "variable reference with code block",
			input: "Set {var}`pkgs.hello` in your `configuration.nix` file.",
			expected: "<p>Set <var>pkgs.hello</var> " +
				"in your <code>configuration.nix</code> file.</p>\n",
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

func TestVarNodeKind(t *testing.T) {
	n := &Node{}
	if kind := n.Kind(); kind != KindVar {
		t.Errorf("Expected node kind %v, got %v", KindVar, kind)
	}
	if !n.Inline() {
		t.Error("Expected node to be inline")
	}
}
