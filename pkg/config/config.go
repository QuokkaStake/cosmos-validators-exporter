package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog"

	"github.com/BurntSushi/toml"
	"github.com/creasty/defaults"
)

type Config struct {
	LogConfig     LogConfig     `toml:"log"`
	TracingConfig TracingConfig `toml:"tracing"`
	ListenAddress string        `default:":9560" toml:"listen-address"`
	Timeout       int           `default:"10"    toml:"timeout"`
	Chains        []*Chain      `toml:"chains"`
}

type LogConfig struct {
	LogLevel   string `default:"info"  toml:"level"`
	JSONOutput bool   `default:"false" toml:"json"`
}

func (c *Config) Validate() error {
	if err := c.TracingConfig.Validate(); err != nil {
		return fmt.Errorf("error in tracing config: %s", err)
	}

	if len(c.Chains) == 0 {
		return errors.New("no chains provided")
	}

	for index, chain := range c.Chains {
		if err := chain.Validate(); err != nil {
			return fmt.Errorf("error in chain %d: %s", index, err)
		}
	}

	return nil
}

func (c *Config) DisplayWarnings(logger *zerolog.Logger) {
	for _, chain := range c.Chains {
		warnings := chain.DisplayWarnings()

		for _, warning := range warnings {
			entry := logger.Warn()
			for label, value := range warning.Labels {
				entry = entry.Str(label, value)
			}

			entry.Msg(warning.Message)
		}
	}
}

func (c *Config) GetCoingeckoCurrencies() []string {
	currencies := []string{}

	for _, chain := range c.Chains {
		for _, denom := range chain.Denoms {
			if denom.CoingeckoCurrency != "" {
				currencies = append(currencies, denom.CoingeckoCurrency)
			}
		}
	}

	return currencies
}

func GetConfig(path string) (*Config, error) {
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	configString := string(configBytes)

	configStruct := Config{}
	if _, err = toml.Decode(configString, &configStruct); err != nil {
		return nil, err
	}

	if err = defaults.Set(&configStruct); err != nil {
		return nil, err
	}

	return &configStruct, nil
}
