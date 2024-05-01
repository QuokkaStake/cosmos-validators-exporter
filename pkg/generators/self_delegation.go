package generators

import (
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type SelfDelegationGenerator struct {
}

func NewSelfDelegationGenerator() *SelfDelegationGenerator {
	return &SelfDelegationGenerator{}
}

func (g *SelfDelegationGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	dataRaw, ok := state.Get(constants.FetcherNameSelfDelegation)
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

	data, _ := dataRaw.(fetchersPkg.SelfDelegationData)

	for chain, delegations := range data.Delegations {
		for validator, delegation := range delegations {
			selfDelegatedTokensGauge.With(prometheus.Labels{
				"chain":   chain,
				"address": validator,
				"denom":   delegation.Denom,
			}).Set(delegation.Amount)
		}
	}

	return []prometheus.Collector{selfDelegatedTokensGauge}
}
