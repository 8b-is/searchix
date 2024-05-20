package config

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
	FetchTimeout  Duration
	ImportTimeout Duration
	OutputPath    string
	Repo          Repository
}
