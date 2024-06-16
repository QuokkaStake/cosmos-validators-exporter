package config

import (
	"errors"

	"github.com/rs/zerolog"
)

type DenomInfo struct {
	Denom              string `toml:"denom"`
	DenomCoefficient   int64  `default:"1000000"            toml:"denom-coefficient"`
	DisplayDenom       string `toml:"display-denom"`
	CoingeckoCurrency  string `toml:"coingecko-currency"`
	DexScreenerChainID string `toml:"dex-screener-chain-id"`
	DexScreenerPair    string `toml:"dex-screener-pair"`
}

func (d *DenomInfo) Validate() error {
	if d.Denom == "" {
		return errors.New("empty denom name")
	}

	if d.DisplayDenom == "" {
		return errors.New("empty display denom name")
	}

	return nil
}

func (d *DenomInfo) DisplayWarnings(chain *Chain, logger *zerolog.Logger) {
	if d.CoingeckoCurrency == "" && (d.DexScreenerPair == "" || d.DexScreenerChainID == "") {
		logger.Warn().
			Str("chain", chain.Name).
			Str("denom", d.Denom).
			Msg("Currency code not set, not fetching exchange rate.")
	}
}

type DenomInfos []*DenomInfo

func (d DenomInfos) Find(denom string) *DenomInfo {
	for _, info := range d {
		if denom == info.Denom {
			return info
		}
	}

	return nil
}
