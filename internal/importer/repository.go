package importer

import (
	"fmt"
	"strings"
)

type RepoType int

const (
	GitHub = iota + 1
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
		return Unknown, fmt.Errorf("unsupported repo type %s", name)
	}
}

func (f *RepoType) UnmarshalText(text []byte) error {
	var err error
	*f, err = parseRepoType(string(text))

	return err
}
