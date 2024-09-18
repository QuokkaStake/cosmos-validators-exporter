package config

import (
	"errors"
	"fmt"
	"main/pkg/fs"

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

func (c *Config) DisplayWarnings() []Warning {
	warnings := []Warning{}

	for _, chain := range c.Chains {
		warnings = append(warnings, chain.DisplayWarnings()...)
	}

	return warnings
}

func GetConfig(path string, filesystem fs.FS) (*Config, error) {
	configBytes, err := filesystem.ReadFile(path)
	if err != nil {
		return nil, err
	}

	configString := string(configBytes)

	configStruct := Config{}
	if _, err = toml.Decode(configString, &configStruct); err != nil {
		return nil, err
	}

	defaults.MustSet(&configStruct)
	return &configStruct, nil
}
