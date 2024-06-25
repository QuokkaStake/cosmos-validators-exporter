package generators

import (
	"main/pkg/config"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type CommissionGenerator struct {
	Chains []*config.Chain
}

func NewCommissionGenerator(chains []*config.Chain) *CommissionGenerator {
	return &CommissionGenerator{Chains: chains}
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

	for _, chain := range g.Chains {
		chainCommissions, ok := data.Commissions[chain.Name]
		if !ok {
			continue
		}

		for _, validator := range chain.Validators {
			validatorCommissions, ok := chainCommissions[validator.Address]
			if !ok {
				continue
			}

			for _, balance := range validatorCommissions {
				amountConverted := chain.Denoms.Convert(&balance)
				commissionUnclaimedTokens.With(prometheus.Labels{
					"chain":   chain.Name,
					"address": validator.Address,
					"denom":   amountConverted.Denom,
				}).Set(amountConverted.Amount)
			}
		}
	}

	return []prometheus.Collector{commissionUnclaimedTokens}
}
