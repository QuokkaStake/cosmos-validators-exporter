package main

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Manager struct {
	Config    Config
	Coingecko *Coingecko
	Logger    zerolog.Logger
}

func NewManager(config Config, logger *zerolog.Logger) *Manager {
	return &Manager{
		Config:    config,
		Coingecko: NewCoingecko(logger),
		Logger:    logger.With().Str("component", "manager").Logger(),
	}
}

func (m *Manager) GetAllValidators() []ValidatorQuery {
	currenciesList := m.Config.GetCoingeckoCurrencies()
	currenciesRates := m.Coingecko.FetchPrices(currenciesList)

	length := 0
	for _, chain := range m.Config.Chains {
		for range chain.Validators {
			length++
		}
	}

	validators := make([]ValidatorQuery, length)

	var wg sync.WaitGroup
	wg.Add(length)

	index := 0

	for _, chain := range m.Config.Chains {
		rpc := NewRPC(chain.LCDEndpoint, m.Logger)

		for _, address := range chain.Validators {
			go func(address string, chain Chain, index int) {
				defer wg.Done()

				start := time.Now()

				query := ValidatorQuery{
					Chain:   chain.Name,
					Address: address,
				}

				info, err := rpc.GetValidator(address)
				query.Duration = time.Since(start)

				if err != nil {
					m.Logger.Error().
						Err(err).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying validator")
					query.Success = false
					validators[index] = query
					return
				}

				query.Success = true

				infoConverted := NewValidatorInfo(info.Validator)

				price, hasPrice := currenciesRates[chain.CoingeckoCurrency]
				if hasPrice {
					infoConverted.TokensUSD = m.CalculatePrice(infoConverted.Tokens, price, chain.DenomCoefficient)
				}

				if delegators, err := rpc.GetDelegationsCount(address); err != nil {
					m.Logger.Error().
						Err(err).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying validator delegations count")
				} else {
					infoConverted.DelegatorsCount = StrToInt64(delegators.Pagination.Total)
				}

				query.Info = &infoConverted

				validators[index] = query
			}(address, chain, index)

			index++
		}
	}

	wg.Wait()

	return validators
}

// func (m *Manager) MaybeGetUsdPrice(
// 	chain Chain,
// 	balances Balances,
// 	rates map[string]float64,
// ) float64 {
// 	price, hasPrice := rates[chain.CoingeckoCurrency]
// 	if !hasPrice {
// 		return 0
// 	}

// 	var usdPriceTotal float64 = 0
// 	for _, balance := range balances {
// 		if balance.Denom == chain.BaseDenom {
// 			usdPriceTotal += m.CalculatePrice(balance)
// 			usdPriceTotal += StrToFloat64(balance.Amount) * price / float64(chain.DenomCoefficient)
// 		}
// 	}

// 	return usdPriceTotal
// }

func (m *Manager) CalculatePrice(amount float64, rate float64, coefficient int64) float64 {
	return amount * rate / float64(coefficient)
}
