package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"

	"searchix/internal/config"
	"searchix/internal/server"
)

var (
	liveReload = flag.Bool("live", false, "whether to enable live reloading (development)")
	configFile = flag.String("config", "config.toml", "config file to use")
)

func main() {
	flag.Parse()

	conf, err := config.GetConfig(*configFile)
	if err != nil {
		log.Panicf("error parsing configuration file: %v", err)
	}
	slog.SetLogLoggerLevel(conf.LogLevel)
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("searchix: ")

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt)
	sv, err := server.New(conf, *liveReload)
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
