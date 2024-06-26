package generators

import (
	"main/pkg/config"
	"main/pkg/constants"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type IsConsumerGenerator struct {
	Chains []*config.Chain
}

func NewIsConsumerGenerator(chains []*config.Chain) *IsConsumerGenerator {
	return &IsConsumerGenerator{Chains: chains}
}

func (g *IsConsumerGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	isConsumerGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "is_consumer",
			Help: "Whether the chain is consumer (1 if yes, 0 if no)",
		},
		[]string{"chain"},
	)

	for _, chain := range g.Chains {
		isConsumerGauge.With(prometheus.Labels{
			"chain": chain.Name,
		}).Set(0)

		for _, consumerChain := range chain.ConsumerChains {
			isConsumerGauge.With(prometheus.Labels{
				"chain": consumerChain.Name,
			}).Set(1)
		}
	}

	return []prometheus.Collector{isConsumerGauge}
}
