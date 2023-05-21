package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/mcuadros/go-defaults"
)

type Validator struct {
	Address          string `toml:"address"`
	ConsensusAddress string `toml:"consensus-address"`
}

func (v *Validator) Validate() error {
	if v.Address == "" {
		return fmt.Errorf("validator address is expected!")
	}

	return nil
}

type Chain struct {
	Name               string          `toml:"name"`
	LCDEndpoint        string          `toml:"lcd-endpoint"`
	CoingeckoCurrency  string          `toml:"coingecko-currency"`
	DexScreenerChainID string          `toml:"dex-screener-chain-id"`
	DexScreenerPair    string          `toml:"dex-screener-pair"`
	BaseDenom          string          `toml:"base-denom"`
	Denom              string          `toml:"denom"`
	DenomCoefficient   int64           `toml:"denom-coefficient" default:"1000000"`
	BechWalletPrefix   string          `toml:"bech-wallet-prefix"`
	Validators         []Validator     `toml:"validators"`
	Queries            map[string]bool `toml:"queries"`
}

func (c *Chain) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("empty chain name")
	}

	if c.LCDEndpoint == "" {
		return fmt.Errorf("no LCD endpoint provided")
	}

	if len(c.Validators) == 0 {
		return fmt.Errorf("no validators provided")
	}

	for index, validator := range c.Validators {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("error in validator #%d: %s", index, err)
		}
	}

	return nil
}

func (c *Chain) QueryEnabled(query string) bool {
	if value, ok := c.Queries[query]; !ok {
		return true // all queries are enabled by default
	} else {
		return value
	}
}

type Config struct {
	LogConfig     LogConfig `toml:"log"`
	ListenAddress string    `toml:"listen-address" default:":9550"`
	Timeout       int       `toml:"timeout" default:"10"`
	Chains        []Chain   `toml:"chains"`
}

type LogConfig struct {
	LogLevel   string `toml:"level" default:"info"`
	JSONOutput bool   `toml:"json" default:"false"`
}

func (c *Config) Validate() error {
	if len(c.Chains) == 0 {
		return fmt.Errorf("no chains provided")
	}

	for index, chain := range c.Chains {
		if err := chain.Validate(); err != nil {
			return fmt.Errorf("error in chain %d: %s", index, err)
		}
	}

	return nil
}

func (c *Config) GetCoingeckoCurrencies() []string {
	currencies := []string{}

	for _, chain := range c.Chains {
		if chain.CoingeckoCurrency != "" {
			currencies = append(currencies, chain.CoingeckoCurrency)
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

	defaults.SetDefaults(&configStruct)
	return &configStruct, nil
}
