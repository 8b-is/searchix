package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"

	"badc0de.net/pkg/flagutil"

	"go.alanpearce.eu/searchix/internal/config"
	"go.alanpearce.eu/searchix/internal/importer"
	"go.alanpearce.eu/searchix/web"
	"go.alanpearce.eu/x/log"
)

var (
	configFile         = flag.String("config", "config.toml", "config `file` to use")
	printDefaultConfig = flag.Bool(
		"print-default-config",
		false,
		"print default configuration and exit",
	)
	dev        = flag.Bool("dev", false, "enable live reloading and nicer logging")
	replace    = flag.Bool("replace", false, "replace existing index and exit")
	update     = flag.Bool("update", false, "update index and exit")
	version    = flag.Bool("version", false, "print version information")
	cpuprofile = flag.String("cpuprofile", "", "enable CPU profiling and save to `file`")
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
	if *cpuprofile != "" {
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

	s, err := web.New(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialise searchix", "error", err)
	}

	err = importer.SetupIndex(ctx, cfg, &importer.Options{
		Update:    *update,
		Replace:   *replace,
		LowMemory: cfg.Importer.LowMemory,
		Logger:    logger,
	})
	if err != nil {
		logger.Fatal("Failed to setup index", "error", err)
	}

	if *replace || *update {
		return
	}

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
