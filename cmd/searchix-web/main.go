package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"

	flags "github.com/jessevdk/go-flags"

	"alin.ovh/searchix/frontend"
	"alin.ovh/searchix/internal/components"
	"alin.ovh/searchix/internal/config"
	"alin.ovh/searchix/internal/file"
	"alin.ovh/searchix/internal/importer"
	"alin.ovh/searchix/internal/index"
	"alin.ovh/searchix/internal/manpages"
	"alin.ovh/searchix/internal/server"
	"alin.ovh/searchix/internal/storage"
	"alin.ovh/searchix/web"
	"alin.ovh/x/log"
)

var options struct {
	Config             flags.Filename `long:"config" default:"config.toml" description:"config file to use"`
	PrintDefaultConfig bool           `long:"print-default-config" description:"print default configuration and exit"`
	GenerateErrorPage  bool           `long:"generate-error-page" description:"generate error page and exit"`
	Dev                bool           `long:"dev" description:"enable live reloading and nicer logging"`
	Prefetch           bool           `long:"prefetch" description:"pre-fetch data and exit"`
	Replace            bool           `long:"replace" description:"replace existing storage and exit"`
	Reindex            bool           `long:"reindex" description:"reindex existing index and exit"`
	Rebuild            bool           `long:"rebuild" description:"rebuild existing index and exit"`
	Offline            bool           `long:"offline" description:"run in offline mode"`
	Version            bool           `long:"version" description:"print version information"`
	CPUProfile         flags.Filename `long:"cpuprofile" description:"output CPU profile to FILE" value-name:"FILE"`
}

var parser = flags.NewParser(&options, flags.Default)

func main() {
	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
	if flags.WroteHelp(err) {
		return
	}

	if options.Version {
		_, err := fmt.Fprintf(os.Stderr, "searchix %s\n", config.Version)
		if err != nil {
			panic("can't write to standard error?!")
		}
		os.Exit(0)
	}
	if options.PrintDefaultConfig {
		_, err := fmt.Print(config.GetDefaultConfig())
		if err != nil {
			panic("can't write to standard output?!")
		}
		os.Exit(0)
	}
	if options.GenerateErrorPage {
		assets, err := frontend.New()
		if err != nil {
			panic("failed to create assets: " + err.Error())
		}

		err = components.ErrorTemplate(components.TemplateData{
			Source:        nil,
			Sources:       []*config.Source{},
			Query:         "",
			ExtraHeadHTML: "",
			Code:          0,
			Message:       `{{placeholder "http.error.status_code"}} {{placeholder "http.error.status_text"}}`,
			Assets:        assets,
		}).Render(os.Stdout)
		if err != nil {
			panic("failed to render error template: " + err.Error())
		}
		os.Exit(0)
	}

	if options.CPUProfile != "" {
		//nolint:forbidigo // admin specifies profile file location
		f, err := os.Create(string(options.CPUProfile))
		if err != nil {
			panic("can't create CPU profile: " + err.Error())
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			panic("can't start CPU profile: " + err.Error())
		}
		defer pprof.StopCPUProfile()
	}

	logger := log.Configure(!options.Dev)

	cfg, err := config.GetConfig(string(options.Config), logger)
	if err != nil {
		logger.Fatal("Failed to parse config file", "error", err)
	}

	log.SetLevel(cfg.LogLevel)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	root, err := file.CreateAndOpenRoot(cfg.DataPath)
	if err != nil {
		logger.Fatal("Failed to open data root", "error", err)
	}
	defer root.Close()

	store, err := storage.New(&storage.Options{
		Root:   root,
		Logger: logger.Named("store"),
	})
	if err != nil {
		logger.Fatal("Failed to create store", "error", err)
	}
	defer store.Close()

	read, write, exists, err := index.OpenOrCreate(
		&index.Options{
			Config:    cfg,
			Force:     options.Reindex,
			LowMemory: cfg.Importer.LowMemory,
			BatchSize: cfg.Importer.BatchSize,
			Logger:    logger.Named("index"),
			Root:      root,
			Store:     store,
		},
	)
	if err != nil {
		logger.Fatal("Failed to open or create index", "error", err)
	}

	mdb := manpages.New(&manpages.Options{
		Logger: logger.Named("manpages"),
		Root:   root,
	})

	s, err := web.New(cfg, logger, &server.Options{
		ReadIndex:      read,
		ManpagesURLMap: mdb,
		Store:          store,
	})
	if err != nil {
		logger.Fatal("Failed to initialise searchix-web", "error", err)
	}

	imp, err := importer.New(cfg, &importer.Options{
		Storage:    store,
		WriteIndex: write,
		LowMemory:  cfg.Importer.LowMemory,
		Logger:     logger.Named("importer"),
		Manpages:   mdb,
		Root:       root,
		Offline:    options.Offline,
	})
	if err != nil {
		logger.Fatal("Failed to create importer", "error", err)
	}

	if !exists || options.Replace || options.Prefetch {
		err := imp.Start(ctx, true, options.Prefetch, nil)
		if err != nil {
			logger.Fatal("Failed to start importer", "error", err)
		}

		if options.Replace || options.Prefetch {
			return
		}
	}

	if !exists || options.Reindex {
		for _, source := range cfg.Importer.Sources {
			hadErrors, err := importer.ImportSource(
				ctx,
				logger.Named("importer"),
				store,
				source,
				write,
			)
			if err != nil {
				logger.Fatal("Failed to import source", "source", source.Name, "error", err)
			}

			if hadErrors {
				logger.Warn("Imported source encountered errors", "source", source.Name)
			}
		}

		if options.Reindex {
			return
		}
	}

	err = imp.EnsureSourcesIndexed(ctx, read)
	if err != nil {
		logger.Fatal("Failed to setup index", "error", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		sCtx, cancel := context.WithCancel(ctx)
		defer wg.Done()
		defer cancel()
		err := s.Start(sCtx, options.Dev)
		if err != nil && !errors.Is(err, context.Canceled) {
			// Error starting or closing listener:
			logger.Fatal("error", "error", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		imp.StartUpdateTimer(ctx)
	}()

	<-ctx.Done()
	s.Stop()
	wg.Wait()
}
