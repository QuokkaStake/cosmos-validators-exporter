package config

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog"
)

type Chain struct {
	Name             string      `toml:"name"`
	LCDEndpoint      string      `toml:"lcd-endpoint"`
	BaseDenom        string      `toml:"base-denom"`
	Denoms           DenomInfos  `toml:"denoms"`
	BechWalletPrefix string      `toml:"bech-wallet-prefix"`
	Validators       []Validator `toml:"validators"`
	Queries          Queries     `toml:"queries"`

	ConsumerChains []*ConsumerChain `toml:"consumers"`
}

func (c *Chain) GetQueries() Queries {
	return c.Queries
}

func (c *Chain) GetHost() string {
	return c.LCDEndpoint
}

func (c *Chain) GetName() string {
	return c.Name
}

func (c *Chain) Validate() error {
	if c.Name == "" {
		return errors.New("empty chain name")
	}

	if c.LCDEndpoint == "" {
		return errors.New("no LCD endpoint provided")
	}

	if len(c.Validators) == 0 {
		return errors.New("no validators provided")
	}

	for index, validator := range c.Validators {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("error in validator #%d: %s", index, err)
		}
	}

	for index, denomInfo := range c.Denoms {
		if err := denomInfo.Validate(); err != nil {
			return fmt.Errorf("error in denom #%d: %s", index, err)
		}
	}

	for index, chain := range c.ConsumerChains {
		if err := chain.Validate(); err != nil {
			return fmt.Errorf("error in consumer chain #%d: %s", index, err)
		}
	}

	return nil
}

func (c *Chain) DisplayWarnings(logger *zerolog.Logger) {
	if c.BaseDenom == "" {
		logger.Warn().
			Str("chain", c.Name).
			Msg("Base denom is not set")
	}

	for _, denom := range c.Denoms {
		denom.DisplayWarnings(c, logger)
	}
}
