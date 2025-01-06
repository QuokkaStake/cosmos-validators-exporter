package generators

import (
	"github.com/prometheus/client_golang/prometheus"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
)

type UnbondsGenerator struct {
}

func NewUnbondsGenerator() *UnbondsGenerator {
	return &UnbondsGenerator{}
}

func (g *UnbondsGenerator) Generate(state fetchersPkg.State) []prometheus.Collector {
	data, ok := fetchersPkg.StateGet[fetchersPkg.UnbondsData](state, constants.FetcherNameUnbonds)
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
