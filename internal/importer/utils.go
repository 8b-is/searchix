package importer

import (
	"fmt"
	"net/url"
	"searchix/internal/options"

	"github.com/bcicen/jstream"
)

func ValueTypeToString(valueType jstream.ValueType) string {
	switch valueType {
	case jstream.Unknown:
		return "unknown"
	case jstream.Null:
		return "null"
	case jstream.String:
		return "string"
	case jstream.Number:
		return "number"
	case jstream.Boolean:
		return "boolean"
	case jstream.Array:
		return "array"
	case jstream.Object:
		return "object"
	}

	return "very strange"
}

func makeGitHubFileURL(userRepo string, ref string, subPath string, line string) string {
	if ref == "" {
		ref = "master"
	}
	url, _ := url.JoinPath("https://github.com/", userRepo, "blob", ref, subPath)
	if line != "" {
		url = url + "#L" + line
	}

	return url
}

// make configurable?
var channelRepoMap = map[string]string{
	"nixpkgs":      "NixOS/nixpkgs",
	"nix-darwin":   "LnL7/nix-darwin",
	"home-manager": "nix-community/home-manager",
}

func MakeChannelLink(channel string, ref string, subPath string) (*options.Link, error) {
	if channelRepoMap[channel] == "" {
		return nil, fmt.Errorf("don't know what repository relates to channel <%s>", channel)
	}

	return &options.Link{
		Name: fmt.Sprintf("<%s/%s>", channel, subPath),
		URL:  makeGitHubFileURL(channelRepoMap[channel], ref, subPath, ""),
	}, nil
}
