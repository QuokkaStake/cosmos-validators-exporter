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
	dataRaw, ok := state.Get(constants.FetcherNameDelegations)
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

	data, _ := dataRaw.(fetchersPkg.DelegationsData)

	for chain, commissions := range data.Delegations {
		for validator, delegations := range commissions {
			delegationsCountGauge.With(prometheus.Labels{
				"chain":   chain,
				"address": validator,
			}).Set(float64(delegations))
		}
	}

	return []prometheus.Collector{delegationsCountGauge}
}
