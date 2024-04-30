package generators

import (
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type CommissionGenerator struct {
}

func NewCommissionGenerator() *CommissionGenerator {
	return &CommissionGenerator{}
}

func (g *CommissionGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	dataRaw, ok := state.Get(constants.FetcherNameCommission)
	if !ok {
		return []prometheus.Collector{}
	}

	commissionUnclaimedTokens := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "unclaimed_commission",
			Help: "Validator's unclaimed commission (in tokens)",
		},
		[]string{"chain", "address", "denom"},
	)

	data, _ := dataRaw.(fetchersPkg.CommissionData)

	for chain, commissions := range data.Commissions {
		for validator, commission := range commissions {
			for _, balance := range commission {
				commissionUnclaimedTokens.With(prometheus.Labels{
					"chain":   chain,
					"address": validator,
					"denom":   balance.Denom,
				}).Set(balance.Amount)
			}
		}
	}

	return []prometheus.Collector{commissionUnclaimedTokens}
}
