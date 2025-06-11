package optlink

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark"
)

func TestOptLinkExtension(t *testing.T) {
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
			name:  "basic option link",
			input: "[](#opt-systemd.sysusers.enable)",
			expected: "<p><a class=\"option\" href=\"/?query=systemd.sysusers.enable\">" +
				"systemd.sysusers.enable</a></p>\n",
		},
		{
			name:  "option link in sentence",
			input: "You can enable this with [](#opt-systemd.sysusers.enable) in your configuration.",
			expected: "<p>You can enable this with <a class=\"option\" href=\"/?query=systemd.sysusers.enable\">" +
				"systemd.sysusers.enable</a> in your configuration.</p>\n",
		},
		{
			name:  "multiple option links",
			input: "Both [](#opt-systemd.sysusers.enable) and [](#opt-networking.firewall.enable) can be used.",
			expected: "<p>Both <a class=\"option\" href=\"/?query=systemd.sysusers.enable\">" +
				"systemd.sysusers.enable</a> and <a class=\"option\" href=\"/?query=networking.firewall.enable\">" +
				"networking.firewall.enable</a> can be used.</p>\n",
		},
		{
			name:     "not an option link - normal link",
			input:    "[regular link](https://example.com)",
			expected: "<p><a href=\"https://example.com\">regular link</a></p>\n",
		},
		{
			name:     "not an option link - empty link with different format",
			input:    "[](#other-prefix)",
			expected: "<p><a href=\"#other-prefix\"></a></p>\n",
		},
		{
			name:  "option link with special characters",
			input: "[](#opt-users.users.<n>.hashedPassword)",
			expected: "<p><a class=\"option\" href=\"/?query=users.users.%3Cn%3E.hashedPassword\">" +
				"users.users.&lt;n&gt;.hashedPassword</a></p>\n",
		},
		{
			name:  "option link with normal link in sentence",
			input: "Use [](#opt-networking.firewall.enable) and see [the docs](https://nixos.org) for more information.",
			expected: "<p>Use <a class=\"option\" href=\"/?query=networking.firewall.enable\">" +
				"networking.firewall.enable</a> and see <a href=\"https://nixos.org\">the docs</a> for more information.</p>\n",
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

func TestOptLinkNodeKind(t *testing.T) {
	n := &Node{}
	if kind := n.Kind(); kind != KindOptLink {
		t.Errorf("Expected node kind %v, got %v", KindOptLink, kind)
	}
	if !n.Inline() {
		t.Error("Expected node to be inline")
	}
}
