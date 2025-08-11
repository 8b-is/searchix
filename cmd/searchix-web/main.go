package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/pprof"

	"alin.ovh/x/log"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/jessevdk/go-flags"

	"alin.ovh/searchix/internal/config"
)

type Options struct {
	Dev        bool           `long:"dev"        description:"enable live reloading and nicer logging"`
	Config     flags.Filename `long:"config"     description:"config file to use"`
	CPUProfile flags.Filename `long:"cpuprofile" description:"output CPU profile to FILE"              value-name:"FILE"`
}

var (
	globalOptions Options
	cfg           *config.Config
	logger        *log.Logger
)

var parser = flags.NewParser(&globalOptions, flags.HelpFlag|flags.PassDoubleDash)

func main() {
	parser.CommandHandler = func(cmd flags.Commander, args []string) (err error) {
		switch cmd.(type) {
		case *PrintDefaultsOptions, *Version, *CheckConfig:
		default:
			logger = log.Configure(!globalOptions.Dev)

			cfg, err = config.GetConfig(string(globalOptions.Config), logger.Named("config"))
			if err != nil {
				return fault.Wrap(err, fmsg.With("Failed to parse config file"))
			}

			log.SetLevel(cfg.LogLevel)
		}

		if globalOptions.CPUProfile != "" {
			//nolint:forbidigo // admin specifies profile file location
			f, err := os.Create(string(globalOptions.CPUProfile))
			if err != nil {
				panic("can't create CPU profile: " + err.Error())
			}
			err = pprof.StartCPUProfile(f)
			if err != nil {
				panic("can't start CPU profile: " + err.Error())
			}
			defer pprof.StopCPUProfile()
		}

		return cmd.Execute(args)
	}

	_, err := parser.Parse()
	if err != nil {
		var flagErr *flags.Error
		if errors.As(err, &flagErr) {
			switch flagErr.Type {
			case flags.ErrHelp:
				parser.WriteHelp(os.Stdout)

				return
			default:
				fmt.Println(flagErr.Error())
			}
		} else {
			fmt.Println(err)
		}

		os.Exit(1)
	}
}
