package main

import (
	"context"
	"errors"
	"math"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/ecnepsnai/sdnotify"
	"golang.org/x/term"

	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/importer"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/manpages"
	"alin.ovh/searchix/internal/server"
	"alin.ovh/searchix/web"
)

type ServeOptions struct{}

func (opts *ServeOptions) Execute(_ []string) (err error) {
	signals := []os.Signal{os.Interrupt, syscall.SIGTERM}
	if term.IsTerminal(int(os.Stdout.Fd())) {
		signals = append(signals, syscall.SIGHUP)
	}

	root, err := file.CreateAndOpenRoot(cfg.DataPath)
	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to open data root"))
	}
	defer root.Close()

	read, write, err := index.OpenOrCreate(
		&index.Options{
			Config:    cfg,
			LowMemory: cfg.Importer.LowMemory,
			BatchSize: cfg.Importer.BatchSize,
			Logger:    logger.Named("index"),
			Root:      root,
		},
	)
	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to open or create index"))
	}

	mdb := manpages.New(&manpages.Options{
		Logger: logger.Named("manpages"),
		Root:   root,
	})

	s, err := web.New(cfg, logger, &server.Options{
		ReadIndex:      read,
		ManpagesURLMap: mdb,
	})
	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to initialise searchix-web"))
	}

	imp, err := importer.New(cfg, &importer.Options{
		ReadIndex:  read,
		WriteIndex: write,
		LowMemory:  cfg.Importer.LowMemory,
		Logger:     logger.Named("importer"),
		Manpages:   mdb,
		Root:       root,
	})
	if err != nil {
		return fault.Wrap(err, fmsg.With("Failed to create importer"))
	}

	ctx, stop := signal.NotifyContext(context.Background(), signals...)
	defer stop()

	wg := &sync.WaitGroup{}
	wg.Go(func() {
		sCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		defer wg.Done()
		err := s.Start(sCtx, globalOptions.Dev)
		if err != nil && !errors.Is(err, context.Canceled) {
			// Error starting or closing listener:
			logger.Fatal("error", "error", err)
		}
	})

	err = sdnotify.Ready()
	if err != nil {
		logger.Fatal("failed to notify systemd of readiness", "error", err)
	}

	reimport := make(chan os.Signal, 1)
	go func() {
		for sig := range reimport {
			err := sdnotify.Status("re-indexing")
			if err != nil {
				logger.Warn("failed to update systemd status", "error", err)
			}

			if sig == syscall.SIGUSR1 {
				logger.Info("manual fetch on SIGUSR1")
				err := imp.Fetch(ctx, true, nil)
				if err != nil {
					logger.Warn("manual fetch error", "error", err)
				}
				logger.Info("manual fetch succeeded")
			}

			logger.Info("manual re-index", "signal", sig.String())
			err = imp.Index(ctx, nil)
			if err != nil {
				logger.Error("manual index error", "error", err)
			}
			logger.Info("manual re-index completed")

			if sig == syscall.SIGUSR1 {
				logger.Info("manual prune")
				err = imp.Prune(ctx)
				if err != nil {
					logger.Error("manual prune error", "error", err)
				}
				logger.Info("manual prune completed")
			}

			err = sdnotify.Status("completed manual re-index")
			if err != nil {
				logger.Warn("failed to update systemd status", "error", err)
			}
		}
	}()
	signal.Notify(reimport, syscall.SIGUSR1, syscall.SIGUSR2)

	wg.Go(func() {
		imp.StartUpdateTimer(ctx)
	})

	if wdu, defined := os.LookupEnv("WATCHDOG_USEC"); defined {
		wd, err := time.ParseDuration(wdu + "us")
		if err != nil {
			logger.Fatal("could not parse WATCHDOG_USEC", "value", wdu)
		}
		tickRate := time.Duration(math.Ceil(wd.Seconds()*0.75)) * time.Second

		wg.Go(func() {
			for range time.Tick(tickRate) {
				err := sdnotify.Watchdog()
				if err != nil {
					logger.Warn("")
				}
			}
		})
	}

	<-ctx.Done()
	stop()

	logger.Debug("stopping server")
	err = sdnotify.Stopping()
	if err != nil {
		logger.Warn("failed to notify systemd of shutdown", "error", err)
	}
	s.Stop()

	logger.Debug("waiting for tasks")
	wg.Wait()

	logger.Info("done")

	return
}

func init() {
	var opts ServeOptions

	cmd, err := parser.AddCommand("serve", "run server", "Serve web interface", &opts)
	if err != nil {
		panic(err)
	}

	cmd.Aliases = []string{"run"}
}
