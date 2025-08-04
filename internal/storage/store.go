package storage

import (
	"context"
	"errors"
	"time"

	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/nix"
	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/gob"
	"go.uber.org/zap"

	"go.etcd.io/bbolt"
)

var BatchSize = 10000

type Options struct {
	Root   *file.Root
	Logger *log.Logger
}

type Store struct {
	db  *storm.DB
	log *log.Logger
}

func New(opts *Options) (*Store, error) {
	//nolint:forbidigo // external package
	path := opts.Root.JoinPath("searchix.bolt")
	bb, err := storm.Open(path,
		storm.Codec(gob.Codec),
		storm.BoltOptions(0o600, &bbolt.Options{
			Timeout: 1 * time.Second,
			Logger: Logger{
				opts.Logger.Named("bolt").GetLogger().Sugar().WithOptions(
					zap.IncreaseLevel(zap.InfoLevel),
				),
			},
		}),
	)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to open database"))
	}

	return &Store{
		db:  bb,
		log: opts.Logger,
	}, nil
}

func (s *Store) Close() error {
	err := s.db.Close()
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to close database"))
	}

	return nil
}

func (s *Store) Import(
	ctx context.Context,
	source *config.Source,
	objects <-chan nix.Importable,
) <-chan error {
	errs := make(chan error)
	node := s.db.From(source.Key).WithBatch(true)

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
				s.log.Info("imported", "count", i)
			}
		}

		if err := tx.Commit(); err != nil {
			errs <- fault.Wrap(err, fmsg.With("failed to commit transaction"))
		}
	}()

	return errs
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

	node := s.db.From(source.Key)

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
