package config

import (
	"log/slog"
	"net/url"
	"os"

	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
)

type URL struct {
	*url.URL
}

func (u *URL) UnmarshalText(text []byte) (err error) {
	u.URL, err = url.Parse(string(text))

	if err != nil {
		return errors.WithMessage(err, "could not parse URL")
	}

	return nil
}

type Config struct {
	BaseURL  URL    `toml:"base_url"`
	DataPath string `toml:"data_path"`
	CSP      *CSP   `toml:"content-security-policy"`
	Headers  map[string]string
}

func GetConfig() (*Config, error) {
	config := Config{}
	slog.Debug("reading config.toml")
	file, err := os.Open("config.toml")
	if err != nil {
		return nil, errors.WithMessage(err, "error reading configuration")
	}
	dec := toml.NewDecoder(file)
	err = dec.Decode(&config)
	if err != nil {
		var tomlError toml.DecodeError
		if errors.As(err, &tomlError) {
			return nil, errors.WithMessage(err, tomlError.Error())
		}

		return nil, errors.Wrap(err, "config error")
	}

	return &config, nil
}
