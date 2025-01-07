package generators

import (
	"main/pkg/config"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type SelfDelegationGenerator struct {
	Chains []*config.Chain
}

func NewSelfDelegationGenerator(chains []*config.Chain) *SelfDelegationGenerator {
	return &SelfDelegationGenerator{Chains: chains}
}

func (g *SelfDelegationGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	data, ok := statePkg.StateGet[fetchersPkg.SelfDelegationData](state, constants.FetcherNameSelfDelegation)
	if !ok {
		return []prometheus.Collector{}
	}

	selfDelegatedTokensGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "self_delegated",
			Help: "Validator's self delegated amount (in tokens)",
		},
		[]string{"chain", "address", "denom"},
	)

	for _, chain := range g.Chains {
		chainDelegations, ok := data.Delegations[chain.Name]
		if !ok {
			continue
		}

		for _, validator := range chain.Validators {
			validatorSelfDelegation, ok := chainDelegations[validator.Address]
			if !ok {
				continue
			}

			amountConverted := chain.Denoms.Convert(validatorSelfDelegation)
			if amountConverted == nil {
				continue
			}

			selfDelegatedTokensGauge.With(prometheus.Labels{
				"chain":   chain.Name,
				"address": validator.Address,
				"denom":   amountConverted.Denom,
			}).Set(amountConverted.Amount)
		}
	}

	return []prometheus.Collector{selfDelegatedTokensGauge}
}
