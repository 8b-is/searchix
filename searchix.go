package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"time"

	"searchix/internal/config"
	"searchix/internal/importer"
	"searchix/internal/index"
	"searchix/internal/server"

	"github.com/pelletier/go-toml/v2"
)

var buildVersion string

var (
	configFile = flag.String("config", "config.toml", "config `file` to use")
	liveReload = flag.Bool("live", false, "whether to enable live reloading (development)")
	replace    = flag.Bool("replace", false, "whether to replace existing database, if it exists")
	version    = flag.Bool("version", false, "print version information")
)

func nextOccurrenceOfLocalTime(t toml.LocalTime) time.Time {
	now := time.Now()
	dayTime := t
	nextRun := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		dayTime.Hour,
		dayTime.Minute,
		dayTime.Second,
		0,
		time.Local,
	)
	if nextRun.Before(now) {
		return nextRun.AddDate(0, 0, 1)
	}

	return nextRun
}

func main() {
	flag.Parse()
	if *version {
		fmt.Fprintln(os.Stderr, "searchix", buildVersion, config.CommitSHA)
		os.Exit(0)
	}

	conf, err := config.GetConfig(*configFile)
	if err != nil {
		log.Panicf("error parsing configuration file: %v", err)
	}
	slog.SetLogLoggerLevel(conf.LogLevel)
	if conf.Web.Environment == "production" {
		log.SetFlags(0)
	} else {
		log.SetFlags(log.LstdFlags)
	}

	read, write, exists, err := index.OpenOrCreate(conf.DataPath, *replace)
	if err != nil {
		log.Fatalf("Failed to open or create index: %v", err)
	}

	if !exists {
		slog.Info("Index doesn't exist. Starting build job...")
		err = importer.Start(conf, write, *replace)
		if err != nil {
			log.Fatalf("Failed to build index: %v", err)
		}
	}

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt)
	sv, err := server.New(conf, read, *liveReload)
	if err != nil {
		log.Fatalf("error setting up server: %v", err)
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		sig := <-c
		log.Printf("signal captured: %v", sig)
		<-sv.Stop()
		slog.Debug("server stopped")
	}()

	go func() {
		nextRun := nextOccurrenceOfLocalTime(conf.Importer.UpdateAt)
		for {
			slog.Debug("scheduling next run", "next-run", nextRun)
			<-time.After(time.Until(nextRun))
			wg.Add(1)
			slog.Info("updating index")
			err = importer.Start(conf, write, false)
			wg.Done()
			if err != nil {
				slog.Warn("error updating index", "error", err)
			} else {
				slog.Info("update complete")
			}
			nextRun = nextRun.AddDate(0, 0, 1)
		}
	}()

	sErr := make(chan error)
	wg.Add(1)
	go func() {
		defer wg.Done()
		sErr <- sv.Start()
	}()

	err = <-sErr
	if err != nil {
		// Error starting or closing listener:
		log.Fatalf("error: %v", err)
	}
	wg.Wait()
}
