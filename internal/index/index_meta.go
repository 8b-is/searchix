package index

import (
	"encoding/json"
	"os"
	"time"

	"go.alanpearce.eu/searchix/internal/file"
	"go.alanpearce.eu/x/log"

	"gitlab.com/tozd/go/errors"
)

const CurrentSchemaVersion = 3

type SourceMeta struct {
	Updated      time.Time
	Path         string
	Rev          string
	ProgramsPath string
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

func createMeta(path string, log *log.Logger) (*Meta, errors.E) {
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

func openMeta(path string, log *log.Logger) (*Meta, errors.E) {
	exists, err := file.Exists(path)
	if err != nil {
		return nil, errors.WithMessage(err, "could not check for existence of index metadata")
	}
	if !exists {
		return createMeta(path, log)
	}

	j, baseErr := os.ReadFile(path)
	if baseErr != nil {
		return nil, errors.WithMessage(baseErr, "could not open index metadata file")
	}
	meta := Meta{
		path: path,
		log:  log,
	}

	if err := json.Unmarshal(j, &meta.data); err != nil {
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

func (i *Meta) Save() errors.E {
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

func (i *Meta) GetSourceMeta(source string) *SourceMeta {
	sourceMeta := i.Sources[source]
	if sourceMeta == nil {
		return &SourceMeta{}
	}

	return sourceMeta
}

func (i *Meta) SetSourceMeta(source string, meta *SourceMeta) {
	if i.Sources == nil {
		i.Sources = make(map[string]*SourceMeta)
	}
	i.Sources[source] = meta
}

func (i *Meta) LastUpdated() time.Time {
	var last time.Time
	for _, sourceMeta := range i.Sources {
		if sourceMeta.Updated.After(last) {
			last = sourceMeta.Updated
		}
	}

	return last
}
