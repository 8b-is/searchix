package command

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark"
)

func TestCommandExtension(t *testing.T) {
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
			name:     "basic command reference",
			input:    "{command}`git clone`",
			expected: "<p><span class=\"command\"><code>git clone</code></span></p>\n",
		},
		{
			name:     "command reference in sentence",
			input:    "You can use {command}`mkdir -p` to create nested directories.",
			expected: "<p>You can use <span class=\"command\"><code>mkdir -p</code></span> to create nested directories.</p>\n",
		},
		{
			name:  "multiple command references",
			input: "Both {command}`ls -la` and {command}`find . -name \"*.go\"` can be used.",
			expected: "<p>Both <span class=\"command\"><code>ls -la</code></span> and <span class=\"command\">" +
				"<code>find . -name \"*.go\"</code></span> can be used.</p>\n",
		},
		{
			name:     "incomplete command reference - no closing backtick",
			input:    "{command}`missingBacktick",
			expected: "<p>{command}`missingBacktick</p>\n",
		},
		{
			name:     "incomplete command reference - empty command name",
			input:    "{command}``",
			expected: "<p>{command}``</p>\n",
		},
		{
			name:  "command reference with code block",
			input: "{command}`grep -r` is used in `terminal.sh`",
			expected: "<p><span class=\"command\"><code>grep -r</code></span> " +
				"is used in <code>terminal.sh</code></p>\n",
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

func TestCommandNodeKind(t *testing.T) {
	n := &Node{}
	if kind := n.Kind(); kind != KindCommand {
		t.Errorf("Expected node kind %v, got %v", KindCommand, kind)
	}
	if !n.Inline() {
		t.Error("Expected node to be inline")
	}
}
