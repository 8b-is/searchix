package config

import (
	"fmt"
	"strings"
)

type RepoType int

const (
	UnknownRepoType = iota
	GitHub
)

type Repository struct {
	Type     string `default:"github"`
	Owner    string
	Repo     string
	Revision string
}

func (f RepoType) String() string {
	switch f {
	case GitHub:
		return "github"
	default:
		return fmt.Sprintf("RepoType(%d)", f)
	}
}

func parseRepoType(name string) (RepoType, error) {
	switch strings.ToLower(name) {
	case "github":
		return GitHub, nil
	default:
		return UnknownRepoType, fmt.Errorf("unsupported repo type %s", name)
	}
}

func (f *RepoType) UnmarshalText(text []byte) error {
	var err error
	*f, err = parseRepoType(string(text))

	return err
}
