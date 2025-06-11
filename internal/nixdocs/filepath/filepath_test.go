package filepath

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark"
)

func TestFilePathExtension(t *testing.T) {
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
			name:     "basic file path reference",
			input:    "{file}`/etc/passwd`",
			expected: "<p><span class=\"file-path\">/etc/passwd</span></p>\n",
		},
		{
			name:  "file path reference in sentence",
			input: "The configuration file {file}`/etc/fstab` contains file system information.",
			expected: "<p>The configuration file <span class=\"file-path\">/etc/fstab</span>" +
				" contains file system information.</p>\n",
		},
		{
			name:  "multiple file path references",
			input: "Both {file}`/etc/hosts` and {file}`/etc/resolv.conf` define DNS settings.",
			expected: "<p>Both <span class=\"file-path\">/etc/hosts</span> and <span class=\"file-path\">" +
				"/etc/resolv.conf</span> define DNS settings.</p>\n",
		},
		{
			name:     "incomplete file path reference - no closing backtick",
			input:    "{file}`/missing/backtick",
			expected: "<p>{file}`/missing/backtick</p>\n",
		},
		{
			name:     "incomplete file path reference - empty path",
			input:    "{file}``",
			expected: "<p>{file}``</p>\n",
		},
		{
			name:  "file path reference with code block",
			input: "Edit {file}`~/.bashrc` using your favorite `text editor`.",
			expected: "<p>Edit <span class=\"file-path\">~/.bashrc</span> " +
				"using your favorite <code>text editor</code>.</p>\n",
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

func TestFilePathNodeKind(t *testing.T) {
	n := &Node{}
	if kind := n.Kind(); kind != KindFilePath {
		t.Errorf("Expected node kind %v, got %v", KindFilePath, kind)
	}
	if !n.Inline() {
		t.Error("Expected node to be inline")
	}
}
