package config

import (
	"time"
)

type Source struct {
	Name          string
	Key           string
	Enable        bool
	Fetcher       Fetcher
	Importer      ImporterType
	Channel       string
	URL           string
	Attribute     string
	ImportPath    string
	FetchTimeout  time.Duration
	ImportTimeout time.Duration
	OutputPath    string
	Repo          Repository
}
