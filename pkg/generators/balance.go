package generators

import (
	"main/pkg/config"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type BalanceGenerator struct {
	Chains []*config.Chain
}

func NewBalanceGenerator(chains []*config.Chain) *BalanceGenerator {
	return &BalanceGenerator{Chains: chains}
}

func (g *BalanceGenerator) Generate(state statePkg.State) []prometheus.Collector {
	data, ok := statePkg.StateGet[fetchersPkg.BalanceData](state, constants.FetcherNameBalance)
	if !ok {
		return []prometheus.Collector{}
	}

	walletBalanceTokens := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "wallet_balance",
			Help: "Validator's wallet balance (in tokens)",
		},
		[]string{"chain", "address", "denom"},
	)

	for _, chain := range g.Chains {
		for _, consumer := range chain.ConsumerChains {
			consumerBalances, ok := data.Balances[consumer.Name]
			if !ok {
				continue
			}

			for _, validator := range chain.Validators {
				validatorBalances, ok := consumerBalances[validator.Address]
				if !ok {
					continue
				}

				for _, balance := range validatorBalances {
					amountConverted := consumer.Denoms.Convert(&balance)
					if amountConverted == nil {
						continue
					}

					walletBalanceTokens.With(prometheus.Labels{
						"chain":   consumer.Name,
						"address": validator.Address,
						"denom":   amountConverted.Denom,
					}).Set(amountConverted.Amount)
				}
			}
		}

		chainBalances, ok := data.Balances[chain.Name]
		if !ok {
			continue
		}

		for _, validator := range chain.Validators {
			validatorBalances, ok := chainBalances[validator.Address]
			if !ok {
				continue
			}

			for _, balance := range validatorBalances {
				amountConverted := chain.Denoms.Convert(&balance)
				if amountConverted == nil {
					continue
				}

				walletBalanceTokens.With(prometheus.Labels{
					"chain":   chain.Name,
					"address": validator.Address,
					"denom":   amountConverted.Denom,
				}).Set(amountConverted.Amount)
			}
		}
	}

	return []prometheus.Collector{walletBalanceTokens}
}
