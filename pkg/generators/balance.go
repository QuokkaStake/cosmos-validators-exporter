package generators

import (
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type BalanceGenerator struct {
}

func NewBalanceGenerator() *BalanceGenerator {
	return &BalanceGenerator{}
}

func (g *BalanceGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	dataRaw, ok := state.Get(constants.FetcherNameBalance)
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

	data, _ := dataRaw.(fetchersPkg.BalanceData)

	for chain, commissions := range data.Balances {
		for validator, commission := range commissions {
			for _, balance := range commission {
				walletBalanceTokens.With(prometheus.Labels{
					"chain":   chain,
					"address": validator,
					"denom":   balance.Denom,
				}).Set(balance.Amount)
			}
		}
	}

	return []prometheus.Collector{walletBalanceTokens}
}
