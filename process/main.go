package main

import (
	"log"
	"log/slog"
	"os"
	"searchix/internal/options"

	"github.com/ardanlabs/conf/v3"
	"github.com/pkg/errors"
)

type Config struct {
	Input  string `conf:"short:i,required,help:NixOS options file (json)"`
	Output string `conf:"short:o,default:/dev/stdout"`
}

func main() {
	if os.Getenv("DEBUG") != "" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	log.SetFlags(0)

	config := Config{}
	help, err := conf.Parse("", &config)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			log.Fatalln(help)
		}
		log.Fatalf("parsing command line: %v", err)
	}

	err = options.Process(config.Input, config.Output)
	if err != nil {
		log.Fatalf("Error processing file: %v", err)
	}
}
