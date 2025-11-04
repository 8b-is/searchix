package meta

import (
	"encoding/json"
	"time"

	"alin.ovh/searchix/internal/file"

	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
)

const (
	CurrentSchemaVersion = 9
	BaseName             = "meta.json"
)

type SourceMeta struct {
	IndexedAt time.Time
	StoredAt  time.Time
	UpdatedAt time.Time
	Path      string
	Rev       string
}

type data struct {
	SchemaVersion int
	LastImport    struct {
		StartedAt  time.Time
		FinishedAt time.Time
	}
	Sources map[string]*SourceMeta
}

type Meta struct {
	root *file.Root
	log  *log.Logger
	data
}

func (i *Meta) IsSchemaOutdated() bool {
	return i.SchemaVersion < CurrentSchemaVersion
}

func (i *Meta) Save() error {
	i.SchemaVersion = CurrentSchemaVersion
	j, err := json.MarshalIndent(i.data, "", "  ")
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not prepare index metadata for saving"))
	}
	i.log.Debug("saving index metadata", "path", BaseName)
	err = i.root.WriteFile(BaseName, j, 0o600)
	if err != nil {
		return fault.Wrap(err, fmsg.With("could not save index metadata"))
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
	i.Sources[source] = meta
}

func (i *Meta) LastImported() time.Time {
	var last time.Time
	for _, sourceMeta := range i.Sources {
		if sourceMeta.StoredAt.After(last) {
			last = sourceMeta.StoredAt
		}
	}

	return last
}

func (i *Meta) LastUpdated() time.Time {
	var last time.Time
	for _, sourceMeta := range i.Sources {
		if sourceMeta.UpdatedAt.After(last) {
			last = sourceMeta.UpdatedAt
		}
	}

	return last
}

func Create(root *file.Root, log *log.Logger) (*Meta, error) {
	exists, err := root.Exists(BaseName)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("could not check for existence of index metadata"))
	}
	if exists {
		return nil, fault.New("index metadata already exists")
	}

	return &Meta{
		root: root,
		log:  log,
		data: data{
			SchemaVersion: CurrentSchemaVersion,
			Sources:       make(map[string]*SourceMeta),
		},
	}, nil
}

func Open(root *file.Root, log *log.Logger) (*Meta, error) {
	exists, err := root.Exists(BaseName)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("could not check for existence of index metadata"))
	}
	if !exists {
		return Create(root, log)
	}

	j, baseErr := root.ReadFile(BaseName)
	if baseErr != nil {
		return nil, fault.Wrap(baseErr, fmsg.With("could not open index metadata file"))
	}
	meta := Meta{
		root: root,
		log:  log,
	}

	if err := json.Unmarshal(j, &meta.data); err != nil {
		return nil, fault.Wrap(err, fmsg.With("index metadata is corrupt, try replacing the index"))
	}

	return &meta, nil
}
