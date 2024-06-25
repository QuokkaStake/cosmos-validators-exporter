package config

import (
	"errors"
	"main/pkg/types"
	"math"
)

type DenomInfo struct {
	Denom             string `toml:"denom"`
	DenomExponent     int64  `default:"6"               toml:"denom-coefficient"`
	DisplayDenom      string `toml:"display-denom"`
	CoingeckoCurrency string `toml:"coingecko-currency"`
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
	if d.CoingeckoCurrency == "" {
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

func (d DenomInfos) Convert(amount *types.Amount) *types.Amount {
	for _, info := range d {
		if info.Denom == amount.Denom {
			return &types.Amount{
				Amount: amount.Amount / math.Pow(10, float64(info.DenomExponent)),
				Denom:  info.DisplayDenom,
			}
		}
	}

	return amount
}
