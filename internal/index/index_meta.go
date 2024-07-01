package index

import (
	"encoding/json"
	"os"
	"time"

	"go.alanpearce.eu/searchix/internal/file"
	"go.alanpearce.eu/x/log"

	"github.com/pkg/errors"
)

const CurrentSchemaVersion = 2

type SourceMeta struct {
	Updated time.Time
	Path    string
	Rev     string
}

type data struct {
	SchemaVersion int
	Sources       map[string]*SourceMeta
}

type Meta struct {
	path string
	log  *log.Logger
	data
}

func createMeta(path string, log *log.Logger) (*Meta, error) {
	exists, err := file.Exists(path)
	if err != nil {
		return nil, errors.WithMessage(err, "could not check for existence of index metadata")
	}
	if exists {
		return nil, errors.New("index metadata already exists")
	}

	return &Meta{
		path: path,
		log:  log,
		data: data{
			SchemaVersion: CurrentSchemaVersion,
		},
	}, nil
}

func openMeta(path string, log *log.Logger) (*Meta, error) {
	exists, err := file.Exists(path)
	if err != nil {
		return nil, errors.WithMessage(err, "could not check for existence of index metadata")
	}
	if !exists {
		return createMeta(path, log)
	}

	j, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.WithMessage(err, "could not open index metadata file")
	}
	meta := Meta{
		path: path,
		log:  log,
	}
	err = json.Unmarshal(j, &meta.data)
	if err != nil {
		return nil, errors.WithMessage(err, "index metadata is corrupt, try replacing the index")
	}

	meta.checkSchemaVersion()

	return &meta, nil
}

func (i *Meta) checkSchemaVersion() {
	if i.SchemaVersion < CurrentSchemaVersion {
		i.log.Warn(
			"Index schema version out of date, suggest re-indexing",
			"schema_version",
			i.SchemaVersion,
			"latest_version",
			CurrentSchemaVersion,
		)
	}
}

func (i *Meta) Save() error {
	i.SchemaVersion = CurrentSchemaVersion
	j, err := json.Marshal(i.data)
	if err != nil {
		return errors.WithMessage(err, "could not prepare index metadata for saving")
	}
	i.log.Debug("saving index metadata", "path", i.path)
	err = os.WriteFile(i.path, j, 0o600)
	if err != nil {
		return errors.WithMessage(err, "could not save index metadata")
	}

	return nil
}

func (i *Meta) GetSourceMeta(source string) SourceMeta {
	sourceMeta := i.data.Sources[source]
	if sourceMeta == nil {
		return SourceMeta{}
	}

	return *sourceMeta
}

func (i *Meta) SetSourceMeta(source string, meta SourceMeta) {
	if i.data.Sources == nil {
		i.data.Sources = make(map[string]*SourceMeta)
	}
	i.data.Sources[source] = &meta
}
