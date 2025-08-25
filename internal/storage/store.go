package storage

import (
	"context"
	"errors"
	"time"

	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/gob"
	"go.etcd.io/bbolt"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/nix"
)

var BatchSize = 50000

type Options struct {
	Replace   bool
	LowMemory bool
	Root      *file.Root
	Logger    *log.Logger
}

type Store struct {
	*storm.DB
	new bool
	log *log.Logger
}

const filename = "searchix.bolt"

func New(opts *Options) (*Store, error) {
	exists, err := opts.Root.Exists(filename)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to check if file exists"))
	}

	if opts.Replace && exists {
		err = opts.Root.Remove(filename)
		if err != nil {
			return nil, fault.Wrap(err, fmsg.With("failed to remove existing file"))
		}

		exists = false
	}

	//nolint:forbidigo // external package
	path := opts.Root.JoinPath(filename)
	bb, err := storm.Open(path,
		storm.Codec(gob.Codec),
		storm.BoltOptions(0o600, &bbolt.Options{
			FreelistType:   bbolt.FreelistMapType,
			NoFreelistSync: true,
			NoGrowSync:     true,
			Timeout:        1 * time.Second,
		}),
	)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to open database"))
	}

	if !opts.LowMemory {
		bb.Bolt.AllocSize = 256 * 1024 * 1024
	}

	return &Store{
		DB:  bb,
		new: !exists,
		log: opts.Logger,
	}, nil
}

func (s *Store) IsNew() bool {
	return s.new
}

func (s *Store) Close() error {
	err := s.DB.Close()
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to close database"))
	}

	return nil
}

func (s *Store) MakeSourceImporter(
	source *config.Source,
) func(context.Context, <-chan nix.Importable) <-chan error {
	return func(ctx context.Context, objects <-chan nix.Importable) <-chan error {
		errs := make(chan error)
		node := s.DB.From(source.Key).WithBatch(true)

		i := 0

		var save func(storm.Node, nix.Importable) error
		switch source.Importer {
		case config.Packages:
			save = saveGen[nix.Package]
		case config.Options:
			save = saveGen[nix.Option]
		default:
			errs <- fault.New("invalid importer")

			return errs
		}

		go func() {
			defer close(errs)
			tx, err := node.Begin(true)
			if err != nil {
				errs <- fault.Wrap(err, fmsg.With("failed to begin transaction"))

				return
			}
			defer func() {
				if err := tx.Rollback(); err != nil {
					if !errors.Is(err, storm.ErrNotInTransaction) {
						errs <- fault.Wrap(err, fmsg.With("failed to rollback transaction"))
					}
				}
			}()

		outer:
			for obj := range objects {
				i++
				select {
				case <-ctx.Done():
					s.log.Warn("import aborted")

					break outer
				default:
				}

				err := save(tx, obj)
				if err != nil {
					errs <- fault.Wrap(err, fmsg.With("failed to save object"))
				}

				if i%BatchSize == 0 {
					s.log.Debug("imported", "count", i)
					err := tx.Commit()
					if err != nil {
						errs <- fault.Wrap(err, fmsg.With("failed to commit transaction"))
					}
					tx, err = node.Begin(true)
					if err != nil {
						errs <- fault.Wrap(err, fmsg.With("failed to begin transaction"))
					}
				}
			}

			if err := tx.Commit(); err != nil {
				errs <- fault.Wrap(err, fmsg.With("failed to commit transaction"))
			}
		}()

		return errs
	}
}

func saveGen[T nix.Importable](node storm.Node, obj nix.Importable) error {
	doc, ok := obj.(T)
	if !ok {
		return fault.Newf("invalid type: %T", obj)
	}

	if err := node.Save(&doc); err != nil {
		return fault.Wrap(err, fmsg.With("failed to save document"))
	}

	return nil
}

func (s *Store) GetDocument(
	source *config.Source,
	id string,
) (nix.Importable, error) {
	var doc nix.Importable
	var err error

	node := s.From(source.Key)

	switch source.Importer {
	case config.Packages:
		doc = &nix.Package{}
		err = node.One("Attribute", id, doc)
	case config.Options:
		doc = &nix.Option{}
		err = node.One("Name", id, doc)
	default:
		return nil, fault.New("invalid importer type")
	}

	if err != nil {
		return nil, fault.Wrap(
			err,
			fmsg.Withf("failed to get document source: %s id: %s", source.Key, id),
		)
	}

	return doc, nil
}

func MakeSourceExporter[T nix.Importable](
	store *Store,
	source *config.Source,
	batchSize int,
) func(context.Context) (<-chan nix.Importable, <-chan error) {
	return func(_ context.Context) (<-chan nix.Importable, <-chan error) {
		results := make(chan nix.Importable, 1)
		errs := make(chan error)

		go func() {
			defer close(results)
			defer close(errs)

			var obj T
			objs := make([]T, 0, batchSize)
			node := store.From(source.Key)
			count, err := node.Count(&obj)
			if err != nil {
				errs <- fault.Wrap(
					err,
					fmsg.Withf("failed to count documents source: %s", source.Key),
				)

				return
			}

			limit := min(batchSize, count)
			for offset := 0; offset < count; offset += batchSize {
				err := node.All(&objs, storm.Skip(offset), storm.Limit(limit))
				if err != nil {
					errs <- fault.Wrap(
						err,
						fmsg.Withf("failed to export documents source: %s offset: %d limit: %d", source.Key, offset, limit),
					)

					return
				}
				for _, obj := range objs {
					results <- obj
				}
			}
		}()

		return results, errs
	}
}
