package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"

	"searchix/internal/server"

	"github.com/ardanlabs/conf/v3"
	"github.com/pkg/errors"
)

var (
	CommitSHA string
	ShortSHA  string
)

func main() {
	if os.Getenv("DEBUG") != "" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("searchix: ")

	runtimeConfig := server.Config{}
	help, err := conf.Parse("", &runtimeConfig)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			os.Exit(1)
		}
		log.Panicf("parsing runtime configuration: %v", err)
	}

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt)
	sv, err := server.New(&runtimeConfig)
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
	if !runtimeConfig.InDevServer {
		log.Printf("server listening on %s", sv.Addr)
	}

	err = <-sErr
	if err != nil {
		// Error starting or closing listener:
		log.Fatalf("error: %v", err)
	}
	wg.Wait()
}
