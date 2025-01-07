package generators

import (
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type InflationGenerator struct {
}

func NewInflationGenerator() *InflationGenerator {
	return &InflationGenerator{}
}

func (g *InflationGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	data, ok := statePkg.StateGet[fetchersPkg.InflationData](state, constants.FetcherNameInflation)
	if !ok {
		return []prometheus.Collector{}
	}

	inflationGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "inflation",
			Help: "Chain inflation",
		},
		[]string{"chain"},
	)

	for chain, inflation := range data.Inflation {
		inflationGauge.With(prometheus.Labels{
			"chain": chain,
		}).Set(inflation.MustFloat64())
	}

	return []prometheus.Collector{inflationGauge}
}
