package config

import (
	"errors"
	"main/pkg/constants"
	"main/pkg/types"
	"math"

	"github.com/guregu/null/v5"
)

type DenomInfo struct {
	Denom             string    `toml:"denom"`
	DenomExponent     int64     `default:"6"               toml:"denom-exponent"`
	DisplayDenom      string    `toml:"display-denom"`
	CoingeckoCurrency string    `toml:"coingecko-currency"`
	Ignore            null.Bool `default:"false"           toml:"ignore"`
}

func (d *DenomInfo) Validate() error {
	if d.Denom == "" {
		return errors.New("empty denom name")
	}

	if d.DisplayDenom == "" && !d.Ignore.Bool {
		return errors.New("empty display denom name")
	}

	return nil
}

func (d *DenomInfo) DisplayWarnings(chain *Chain) []Warning {
	warnings := []Warning{}
	if d.CoingeckoCurrency == "" && !d.Ignore.Bool {
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

func (d *DenomInfo) PriceFetchers() []constants.PriceFetcherName {
	if d.CoingeckoCurrency != "" {
		return []constants.PriceFetcherName{constants.PriceFetcherNameCoingecko}
	}

	return []constants.PriceFetcherName{}
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
			if info.Ignore.Bool {
				return nil
			}

			return &types.Amount{
				Amount: amount.Amount / math.Pow(10, float64(info.DenomExponent)),
				Denom:  info.DisplayDenom,
			}
		}
	}

	return amount
}
