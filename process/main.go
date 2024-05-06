package main

import (
	"log"
	"log/slog"
	"os"
	"strings"

	"searchix/internal/options"

	"github.com/ardanlabs/conf/v3"
	"github.com/pkg/errors"
)

type Config struct {
	Input        string `conf:"short:i,required,help:NixOS options file (json)"`
	Output       string `conf:"short:o,default:/dev/stdout"`
	Revision     string `conf:"short:r,flag:revision,default:master"`
	RevisionFile string `conf:"short:f,flag:revision-file"`
	Channel      string `conf:"short:c,flag:channel,default:nixpkgs"`
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
	if config.RevisionFile != "" {
		f, err := os.ReadFile(config.RevisionFile)
		if err != nil {
			log.Fatalf("Error reading revision file %s: %v", config.RevisionFile, err)
		}
		config.Revision = strings.TrimSpace(string(f))
	}

	err = options.Process(config.Input, config.Output, config.Channel, config.Revision)
	if err != nil {
		log.Fatalf("Error processing file: %v", err)
	}
}
