package index

import (
	"encoding/json"
	"log/slog"
	"os"
	"searchix/internal/file"

	"github.com/pkg/errors"
)

const CurrentSchemaVersion = 1

type Meta struct {
	path          string
	SchemaVersion int
}

func createMeta(path string) (*Meta, error) {
	exists, err := file.Exists(path)
	if err != nil {
		return nil, errors.WithMessage(err, "could not check for existence of index metadata")
	}
	if exists {
		return nil, errors.New("index metadata already exists")
	}

	return &Meta{
		path:          path,
		SchemaVersion: CurrentSchemaVersion,
	}, nil
}

func openMeta(path string) (*Meta, error) {
	j, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.WithMessage(err, "could not open index metadata file")
	}
	var meta Meta
	err = json.Unmarshal(j, &meta)
	if err != nil {
		return nil, errors.WithMessage(err, "index metadata is corrupt, try replacing the index")
	}

	meta.checkSchemaVersion()

	return &meta, nil
}

func (i *Meta) checkSchemaVersion() {
	if i.SchemaVersion < CurrentSchemaVersion {
		slog.Warn(
			"Index schema version out of date, suggest re-indexing",
			"schema_version",
			i.SchemaVersion,
			"latest_version",
			CurrentSchemaVersion,
		)
	}
}

func (i *Meta) Save() error {
	j, err := json.Marshal(i)
	if err != nil {
		return errors.WithMessage(err, "could not prepare index metadata for saving")
	}
	err = os.WriteFile(i.path, j, 0o600)
	if err != nil {
		return errors.WithMessage(err, "could not save index metadata")
	}

	return nil
}
