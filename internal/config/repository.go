package config

import (
	"fmt"
	"strings"

	"gitlab.com/tozd/go/errors"
)

type RepoType int

const (
	UnknownRepoType = iota
	GitHub
)

type Repository struct {
	Type     RepoType `toml:""  default:"github" comment:"Currently only 'github' is supported."`
	Owner    string
	Repo     string
	Revision string `toml:"-"`
}

func (f RepoType) String() string {
	switch f {
	case GitHub:
		return "github"
	default:
		return fmt.Sprintf("RepoType(%d)", f)
	}
}

func parseRepoType(name string) (RepoType, errors.E) {
	switch strings.ToLower(name) {
	case "github":
		return GitHub, nil
	default:
		return UnknownRepoType, errors.Errorf("unsupported repo type %s", name)
	}
}

func (f *RepoType) UnmarshalText(text []byte) error {
	var err error
	*f, err = parseRepoType(string(text))

	return err
}

func (f *RepoType) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}
