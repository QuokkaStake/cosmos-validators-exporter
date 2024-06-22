package config

import (
	"errors"
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

func (d *DenomInfo) DisplayWarnings(chain *Chain) []Warning {
	warnings := []Warning{}
	if d.CoingeckoCurrency == "" && (d.DexScreenerPair == "" || d.DexScreenerChainID == "") {
		warnings = append(warnings, Warning{
			Message: "Currency code not set, not fetching exchange rate.",
			Labels: map[string]string{
				"chain": chain.Name,
				"denom": d.Denom,
			},
		})
	}

	return warnings
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
