package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"sync"
	"time"

	"searchix/internal/config"
	"searchix/internal/importer"
	"searchix/internal/index"
	"searchix/internal/server"

	"github.com/getsentry/sentry-go"
	"github.com/pelletier/go-toml/v2"
)

var buildVersion string

var (
	configFile         = flag.String("config", "config.toml", "config `file` to use")
	printDefaultConfig = flag.Bool(
		"print-default-config",
		false,
		"print default configuration and exit",
	)
	liveReload = flag.Bool("live", false, "whether to enable live reloading (development)")
	replace    = flag.Bool(
		"replace",
		false,
		"whether to replace existing database, if it exists",
	)
	version = flag.Bool("version", false, "print version information")
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
		fmt.Fprintf(os.Stderr, "searchix %s", buildVersion)
		if buildVersion != config.CommitSHA && buildVersion != config.ShortSHA {
			fmt.Fprintf(os.Stderr, " %s", config.CommitSHA)
		}
		_, err := fmt.Fprint(os.Stderr, "\n")
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
		fmt.Fprintf(os.Stderr, "error parsing configuration file: %v", err)
		os.Exit(1)
	}
	slog.SetLogLoggerLevel(cfg.LogLevel)
	if cfg.Web.Environment == "production" {
		log.SetFlags(0)
	} else {
		log.SetFlags(log.LstdFlags)
	}

	err = sentry.Init(sentry.ClientOptions{
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		Dsn:              cfg.Web.SentryDSN,
		Environment:      cfg.Web.Environment,
	})
	if err != nil {
		slog.Warn("could not initialise sentry", "error", err)
	}

	var i uint
	cfgEnabledSources := make([]string, len(cfg.Importer.Sources))
	for key := range cfg.Importer.Sources {
		cfgEnabledSources[i] = key
		i++
	}
	slices.Sort(cfgEnabledSources)

	read, write, exists, err := index.OpenOrCreate(cfg.DataPath, *replace)
	if err != nil {
		log.Fatalf("Failed to open or create index: %v", err)
	}

	if !exists || *replace {
		slog.Info("Index doesn't exist. Starting build job...")
		err = importer.Start(cfg, write, *replace, nil)
		if err != nil {
			log.Fatalf("Failed to build index: %v", err)
		}
		if *replace {
			return
		}
	} else {
		indexedSources, err := read.GetEnabledSources()
		if err != nil {
			log.Fatalln("failed to get enabled sources from index")
		}
		slices.Sort(indexedSources)
		if !slices.Equal(cfgEnabledSources, indexedSources) {
			newSources := slices.DeleteFunc(slices.Clone(cfgEnabledSources), func(s string) bool {
				return slices.Contains(indexedSources, s)
			})
			retiredSources := slices.DeleteFunc(slices.Clone(indexedSources), func(s string) bool {
				return slices.Contains(cfgEnabledSources, s)
			})
			if len(newSources) > 0 {
				slog.Info("adding new sources", "sources", newSources)
				err := importer.Start(cfg, write, false, &newSources)
				if err != nil {
					log.Fatalf("failed to update index with new sources: %v", err)
				}
			}
			if len(retiredSources) > 0 {
				slog.Info("removing retired sources", "sources", retiredSources)
				for _, s := range retiredSources {
					err := write.DeleteBySource(s)
					if err != nil {
						log.Fatalf("failed to remove retired source %s from index: %v", s, err)
					}
				}
			}
		}
	}

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt)
	sv, err := server.New(cfg, read, *liveReload)
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

	go func(localHub *sentry.Hub) {
		const monitorSlug = "import"
		localHub.WithScope(func(scope *sentry.Scope) {
			scope.SetContext("monitor", sentry.Context{"slug": monitorSlug})
			monitorConfig := &sentry.MonitorConfig{
				Schedule: sentry.IntervalSchedule(1, sentry.MonitorScheduleUnitDay),
				// minutes
				MaxRuntime:    10,
				CheckInMargin: 5,
				Timezone:      time.Local.String(),
			}

			nextRun := nextOccurrenceOfLocalTime(cfg.Importer.UpdateAt)
			for {
				slog.Debug("scheduling next run", "next-run", nextRun)
				<-time.After(time.Until(nextRun))
				wg.Add(1)
				slog.Info("updating index")

				eventID := localHub.CaptureCheckIn(&sentry.CheckIn{
					MonitorSlug: monitorSlug,
					Status:      sentry.CheckInStatusInProgress,
				}, monitorConfig)

				err = importer.Start(cfg, write, false, nil)
				wg.Done()
				if err != nil {
					slog.Warn("error updating index", "error", err)

					localHub.CaptureException(err)
					localHub.CaptureCheckIn(&sentry.CheckIn{
						ID:          *eventID,
						MonitorSlug: monitorSlug,
						Status:      sentry.CheckInStatusError,
					}, monitorConfig)
				} else {
					slog.Info("update complete")

					localHub.CaptureCheckIn(&sentry.CheckIn{
						ID:          *eventID,
						MonitorSlug: monitorSlug,
						Status:      sentry.CheckInStatusOK,
					}, monitorConfig)
				}
				nextRun = nextRun.AddDate(0, 0, 1)
			}
		})
	}(sentry.CurrentHub().Clone())

	sErr := make(chan error)
	wg.Add(1)
	go func() {
		defer wg.Done()
		sErr <- sv.Start()
	}()

	if cfg.Web.Environment == "development" {
		log.Printf("server listening on %s", cfg.Web.BaseURL.String())
	}

	err = <-sErr
	if err != nil {
		// Error starting or closing listener:
		log.Fatalf("error: %v", err)
	}
	sentry.Flush(2 * time.Second)
	wg.Wait()
}
