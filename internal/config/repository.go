package config

import (
	"fmt"
	"net/url"
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

func (r *Repository) GetFileURL(path string, line ...string) (string, errors.E) {
	switch r.Type {
	case GitHub:
		ref := r.Revision
		if ref == "" {
			ref = "master"
		}
		u, err := url.JoinPath("https://github.com/", r.Owner, r.Repo, "blob", ref, path)
		if err != nil {
			return "", errors.Wrap(err, "failed to join path")
		}
		if len(line) > 0 {
			u += fmt.Sprintf("#L%s", line[0])
		}

		return u, nil
	default:
		return "", errors.Errorf(
			"don't know how to generate a repository URL for %s",
			r.Type.String(),
		)
	}
}

func (r *Repository) String() string {
	switch r.Type {
	case GitHub:
		u, err := url.JoinPath("https://github.com/", r.Owner, r.Repo)
		if err != nil {
			panic(err)
		}

		return u
	default:
		panic("need repository string implementation for type " + r.Type.String())
	}
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
