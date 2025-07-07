package config

import (
	"maps"
	"net/url"
	"os"
	"time"

	"alin.ovh/x/log"
	"github.com/creasty/defaults"
	"github.com/pelletier/go-toml/v2"
	"gitlab.com/tozd/go/errors"
)

var (
	Version string
	DevMode bool
)

const MaxResultsShowAll = 10_000

type URL struct {
	*url.URL
}

func (u *URL) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

func (u *URL) UnmarshalText(text []byte) (err error) {
	u.URL, err = url.Parse(string(text))
	if err != nil {
		return errors.WithMessagef(err, "could not parse URL %s", string(text))
	}

	return nil
}

func (u *URL) JoinPath(elems ...string) *URL {
	return &URL{u.URL.JoinPath(elems...)}
}

func (u *URL) AddRawQuery(key, value string) *URL {
	u.RawQuery = key + "=" + value

	return u
}

type Duration struct {
	time.Duration
}

func (d *Duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Duration) UnmarshalText(text []byte) (err error) {
	d.Duration, err = time.ParseDuration(string(text))
	if err != nil {
		return errors.WithMessagef(err, "could not parse duration %s", string(text))
	}

	return nil
}

func mustURL(in string) (u URL) {
	var err error
	u.URL, err = url.Parse(in)
	if err != nil {
		panic(errors.Errorf("URL cannot be parsed: %s", in))
	}

	return u
}

// this type is necessary as nix's `fromTOML` doesn't support TOML date/time formats
type LocalTime struct {
	toml.LocalTime
}

func (t *LocalTime) MarshalText() ([]byte, error) {
	b, err := t.LocalTime.MarshalText()
	if err != nil {
		return nil, errors.WithMessage(err, "could not marshal time value")
	}

	return b, nil
}

func (t *LocalTime) UnmarshalText(in []byte) (err error) {
	err = t.LocalTime.UnmarshalText(in)
	if err != nil {
		return errors.WithMessage(err, "could not parse time value")
	}

	return nil
}

func mustLocalTime(in string) (time LocalTime) {
	err := time.UnmarshalText([]byte(in))
	if err != nil {
		panic(errors.Errorf("Could not parse time: %s", in))
	}

	return
}

func GetConfig(filename string, log *log.Logger) (*Config, errors.E) {
	config := DefaultConfig
	if filename != "" {
		log.Debug("reading config", "filename", filename)
		f, err := os.Open(filename)
		if err != nil {
			return nil, errors.Wrap(err, "reading config failed")
		}
		defer f.Close()

		dec := toml.NewDecoder(f)
		dec.DisallowUnknownFields()
		err = dec.Decode(&config)
		if err != nil {
			var tomlError *toml.DecodeError
			if errors.As(err, &tomlError) {
				return nil, errors.WithMessage(err, tomlError.Error())
			}
			var missingConfigError *toml.StrictMissingError
			if errors.As(err, &missingConfigError) {
				return nil, errors.Errorf("unexpected config: %s", missingConfigError.String())
			}

			return nil, errors.Wrap(err, "config error")
		}
	}

	DevMode = config.Web.Environment == "development"

	config.Web.ContentSecurityPolicy.ScriptSrc = append(
		config.Web.ContentSecurityPolicy.ScriptSrc,
		config.Web.BaseURL.JoinPath("/static/").String(),
	)

	maps.DeleteFunc(config.Importer.Sources, func(_ string, v *Source) bool {
		return !v.Enable
	})

	for k, v := range config.Importer.Sources {
		if v.Key == "" {
			v.Key = k
		}
		if err := defaults.Set(v); err != nil {
			return nil, errors.Wrap(err, "setting defaults failed")
		}
	}

	return &config, nil
}
