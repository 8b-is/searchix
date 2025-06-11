package option

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark"
)

func TestOptionExtension(t *testing.T) {
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
			name:  "basic option reference",
			input: "{option}`initialHashedPassword`",
			expected: "<p><a class=\"option\" href=\"/?query=initialHashedPassword\">" +
				"<code>initialHashedPassword</code></a></p>\n",
		},
		{
			name:  "option reference in sentence",
			input: "You can set {option}`initialHashedPassword` to configure the initial password.",
			expected: "<p>You can set <a class=\"option\" href=\"/?query=initialHashedPassword\">" +
				"<code>initialHashedPassword</code></a> to configure the initial password.</p>\n",
		},
		{
			name:  "multiple option references",
			input: "Both {option}`initialHashedPassword` and {option}`passwordFile` can be used.",
			expected: "<p>Both <a class=\"option\" href=\"/?query=initialHashedPassword\">" +
				"<code>initialHashedPassword</code></a> and <a class=\"option\" href=\"/?query=passwordFile\">" +
				"<code>passwordFile</code></a> can be used.</p>\n",
		},
		{
			name:     "incomplete option reference - no closing backtick",
			input:    "{option}`missingBacktick",
			expected: "<p>{option}`missingBacktick</p>\n",
		},
		{
			name:     "incomplete option reference - empty option name",
			input:    "{option}``",
			expected: "<p>{option}``</p>\n",
		},
		{
			name:  "option reference with code block",
			input: "{option}`config` is set in `config.nix`",
			expected: "<p><a class=\"option\" href=\"/?query=config\"><code>config</code></a> " +
				"is set in <code>config.nix</code></p>\n",
		},
		{
			name:  "option reference with HTML-like content",
			input: "{option}`users.users.<name>.hashedPassword`",
			expected: "<p><a class=\"option\" href=\"/?query=users.users.%3Cname%3E.hashedPassword\">" +
				"<code>users.users.&lt;name&gt;.hashedPassword</code></a></p>\n",
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

func TestOptionNodeKind(t *testing.T) {
	n := &Node{}
	if kind := n.Kind(); kind != KindOption {
		t.Errorf("Expected node kind %v, got %v", KindOption, kind)
	}
	if !n.Inline() {
		t.Error("Expected node to be inline")
	}
}
