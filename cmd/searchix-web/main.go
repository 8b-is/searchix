package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"

	"badc0de.net/pkg/flagutil"

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

var (
	configFile         = flag.String("config", "config.toml", "config `file` to use")
	printDefaultConfig = flag.Bool(
		"print-default-config",
		false,
		"print default configuration and exit",
	)
	generateErrorPage = flag.Bool("generate-error-page", false, "generate error page and exit")
	dev               = flag.Bool("dev", false, "enable live reloading and nicer logging")
	replace           = flag.Bool("replace", false, "replace existing index and exit")
	prefetch          = flag.Bool("fetch", false, "pre-fetch data and exit")
	offline           = flag.Bool("offline", false, "run in offline mode")
	version           = flag.Bool("version", false, "print version information")
	cpuprofile        = flag.String("cpuprofile", "", "enable CPU profiling and save to `file`")
)

func main() {
	flagutil.Parse()
	if *version {
		_, err := fmt.Fprintf(os.Stderr, "searchix %s\n", config.Version)
		if err != nil {
			panic("can't write to standard error?!")
		}
		os.Exit(0)
	}
	if *printDefaultConfig {
		_, err := fmt.Print(config.GetDefaultConfig())
		if err != nil {
			panic("can't write to standard output?!")
		}
		os.Exit(0)
	}
	if *generateErrorPage {
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

	if *cpuprofile != "" {
		//nolint:forbidigo // admin specifies profile file location
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic("can't create CPU profile: " + err.Error())
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			panic("can't start CPU profile: " + err.Error())
		}
		defer pprof.StopCPUProfile()
	}

	logger := log.Configure(!*dev)

	cfg, err := config.GetConfig(*configFile, logger)
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
			Force:     *replace,
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
		Offline:    *offline,
	})
	if err != nil {
		logger.Fatal("Failed to create importer", "error", err)
	}

	if !exists || *replace || *prefetch {
		err := imp.Start(ctx, true, *prefetch, nil)
		if err != nil {
			logger.Fatal("Failed to start importer", "error", err)
		}

		if *replace || *prefetch {
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
		err := s.Start(sCtx, *dev)
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
