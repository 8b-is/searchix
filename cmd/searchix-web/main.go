package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"badc0de.net/pkg/flagutil"

	"go.alanpearce.eu/searchix"
	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/x/log"
)

var (
	configFile         = flag.String("config", "config.toml", "config `file` to use")
	printDefaultConfig = flag.Bool(
		"print-default-config",
		false,
		"print default configuration and exit",
	)
	dev     = flag.Bool("dev", false, "enable live reloading and nicer logging")
	replace = flag.Bool("replace", false, "replace existing index and exit")
	update  = flag.Bool("update", false, "update index and exit")
	version = flag.Bool("version", false, "print version information")
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

	logger := log.Configure(!*dev)

	cfg, err := config.GetConfig(*configFile, logger)
	if err != nil {
		logger.Fatal("Failed to parse config file", "error", err)
	}

	log.SetLevel(cfg.LogLevel)

	s, err := searchix.New(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialise searchix", "error", err)
	}

	err = s.SetupIndex(&searchix.IndexOptions{
		Update:    *update,
		Replace:   *replace,
		LowMemory: cfg.Importer.LowMemory,
		Logger:    logger,
	})
	if err != nil {
		logger.Fatal("Failed to setup index", "error", err)
	}

	if *replace || *update {
		os.Exit(0)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go func() {
		err = s.Start(ctx, *dev)
		if err != nil {
			// Error starting or closing listener:
			logger.Fatal("error", "error", err)
		}
	}()

	<-ctx.Done()
	logger.Debug("calling stop")
	s.Stop()
	logger.Debug("done")
}
