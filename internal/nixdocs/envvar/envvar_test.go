package envvar

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark"
)

func TestEnvExtension(t *testing.T) {
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
			name:     "basic env variable reference",
			input:    "{env}`XDG_DATA_DIRS`",
			expected: "<p><span class=\"envvar\">XDG_DATA_DIRS</span></p>\n",
		},
		{
			name:     "env variable reference in sentence",
			input:    "The environment variable {env}`PATH` contains a list of directories.",
			expected: "<p>The environment variable <span class=\"envvar\">PATH</span> contains a list of directories.</p>\n",
		},
		{
			name:  "multiple env variable references",
			input: "Both {env}`HOME` and {env}`XDG_CONFIG_HOME` define user directories.",
			expected: "<p>Both <span class=\"envvar\">HOME</span> and <span class=\"envvar\">" +
				"XDG_CONFIG_HOME</span> define user directories.</p>\n",
		},
		{
			name:     "incomplete env variable reference - no closing backtick",
			input:    "{env}`MISSING_BACKTICK",
			expected: "<p>{env}`MISSING_BACKTICK</p>\n",
		},
		{
			name:     "incomplete env variable reference - empty variable name",
			input:    "{env}``",
			expected: "<p>{env}``</p>\n",
		},
		{
			name:  "env variable reference with code block",
			input: "Set {env}`GOPATH` in your `~/.bashrc` file.",
			expected: "<p>Set <span class=\"envvar\">GOPATH</span> " +
				"in your <code>~/.bashrc</code> file.</p>\n",
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

func TestEnvNodeKind(t *testing.T) {
	n := &Node{}
	if kind := n.Kind(); kind != KindEnv {
		t.Errorf("Expected node kind %v, got %v", KindEnv, kind)
	}
	if !n.Inline() {
		t.Error("Expected node to be inline")
	}
}
