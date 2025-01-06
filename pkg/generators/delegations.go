package generators

import (
	"github.com/prometheus/client_golang/prometheus"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
)

type DelegationsGenerator struct {
}

func NewDelegationsGenerator() *DelegationsGenerator {
	return &DelegationsGenerator{}
}

func (g *DelegationsGenerator) Generate(state fetchersPkg.State) []prometheus.Collector {
	data, ok := fetchersPkg.StateGet[fetchersPkg.DelegationsData](state, constants.FetcherNameDelegations)
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
