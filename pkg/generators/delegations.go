package generators

import (
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type DelegationsGenerator struct {
}

func NewDelegationsGenerator() *DelegationsGenerator {
	return &DelegationsGenerator{}
}

func (g *DelegationsGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	data, ok := statePkg.StateGet[fetchersPkg.DelegationsData](state, constants.FetcherNameDelegations)
	if !ok {
		return []prometheus.Collector{}
	}

	delegationsCountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "delegations_count",
			Help: "Validator delegations count",
		},
		[]string{"chain", "address"},
	)

	for chain, allDelegations := range data.Delegations {
		for validator, delegations := range allDelegations {
			delegationsCountGauge.With(prometheus.Labels{
				"chain":   chain,
				"address": validator,
			}).Set(float64(delegations))
		}
	}

	return []prometheus.Collector{delegationsCountGauge}
}
