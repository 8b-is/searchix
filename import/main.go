package main

import (
	"flag"
	"log"
	"log/slog"
	"searchix/internal/config"
	"searchix/internal/importer"
)

var (
	replace    = flag.Bool("replace", false, "whether to replace existing database, if it exists")
	configFile = flag.String("config", "config.toml", "config file to use")
)

func main() {
	flag.Parse()

	cfg, err := config.GetConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	slog.SetLogLoggerLevel(cfg.LogLevel)

	err = importer.Start(cfg, *replace)
	if err != nil {
		log.Fatal(err)
	}
}
