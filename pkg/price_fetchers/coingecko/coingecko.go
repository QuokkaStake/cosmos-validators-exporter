package coingecko

import (
	"github.com/rs/zerolog"
	gecko "github.com/superoo7/go-gecko/v3"
)

type Coingecko struct {
	Client *gecko.Client
	Logger zerolog.Logger
}

func NewCoingecko(logger *zerolog.Logger) *Coingecko {
	return &Coingecko{
		Client: gecko.NewClient(nil),
		Logger: logger.With().Str("component", "coingecko").Logger(),
	}
}

func (c *Coingecko) FetchPrices(currencies []string) map[string]float64 {
	result, err := c.Client.SimplePrice(currencies, []string{"USD"})
	if err != nil {
		c.Logger.Error().Err(err).Msg("Could not get rate")
		return map[string]float64{}
	}

	prices := map[string]float64{}

	for currencyKey, currencyValue := range *result {
		for _, baseCurrencyValue := range currencyValue {
			prices[currencyKey] = float64(baseCurrencyValue)
		}
	}

	return prices
}
