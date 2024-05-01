package generators

import (
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type UnbondsGenerator struct {
}

func NewUnbondsGenerator() *UnbondsGenerator {
	return &UnbondsGenerator{}
}

func (g *UnbondsGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	dataRaw, ok := state.Get(constants.FetcherNameUnbonds)
	if !ok {
		return []prometheus.Collector{}
	}

	unbondsCountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "unbonds_count",
			Help: "Validator unbonds count",
		},
		[]string{"chain", "address"},
	)

	data, _ := dataRaw.(fetchersPkg.UnbondsData)

	for chain, commissions := range data.Unbonds {
		for validator, unbonds := range commissions {
			unbondsCountGauge.With(prometheus.Labels{
				"chain":   chain,
				"address": validator,
			}).Set(float64(unbonds))
		}
	}

	return []prometheus.Collector{unbondsCountGauge}
}
