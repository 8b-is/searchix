package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"

	"badc0de.net/pkg/flagutil"

	"searchix"
	"searchix/internal/config"
)

var (
	configFile         = flag.String("config", "config.toml", "config `file` to use")
	printDefaultConfig = flag.Bool(
		"print-default-config",
		false,
		"print default configuration and exit",
	)
	liveReload = flag.Bool("live", false, "whether to enable live reloading (development)")
	replace    = flag.Bool("replace", false, "replace existing index and exit")
	update     = flag.Bool("update", false, "update index and exit")
	version    = flag.Bool("version", false, "print version information")
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

	cfg, err := config.GetConfig(*configFile)
	if err != nil {
		// only use log functions after the config file has been read successfully
		log.Fatalf("Failed to parse config file: %v", err)
	}
	s, err := searchix.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialise searchix: %v", err)
	}

	err = s.SetupIndex(*update, *replace)
	if err != nil {
		log.Fatalf("Failed to setup index: %v", err)
	}

	if *replace || *update {
		os.Exit(0)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go func() {
		err = s.Start(ctx, *liveReload)
		if err != nil {
			// Error starting or closing listener:
			log.Fatalf("error: %v", err)
		}
	}()

	<-ctx.Done()
	slog.Debug("calling stop")
	s.Stop()
	slog.Debug("done")
}
