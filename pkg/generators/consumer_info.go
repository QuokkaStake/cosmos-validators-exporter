package generators

import (
	"main/pkg/config"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type ConsumerInfoGenerator struct {
	Chains []*config.Chain
}

func NewConsumerInfoGenerator(chains []*config.Chain) *ConsumerInfoGenerator {
	return &ConsumerInfoGenerator{Chains: chains}
}

func (g *ConsumerInfoGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	consumerInfosRaw, ok := state.Get(constants.FetcherNameConsumerInfo)
	if !ok {
		return []prometheus.Collector{}
	}

	consumerInfos, _ := consumerInfosRaw.(fetchersPkg.ConsumerInfoData)

	thresholdPercentGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "threshold_percent",
			Help: "Top-N percent threshold for consumer chains.",
		},
		[]string{
			"chain",
			"chain_id",
			"provider",
		},
	)

	minStakeGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "minimal_stake",
			Help: "Minimal stake to be required to sign blocks on this consumer chain",
		},
		[]string{
			"chain",
			"chain_id",
			"provider",
		},
	)

	for _, chain := range g.Chains {
		consumersInfo, ok := consumerInfos.Info[chain.Name]
		if !ok {
			continue
		}

		for _, consumer := range chain.ConsumerChains {
			consumerInfo, ok := consumersInfo[consumer.ChainID]
			if !ok {
				continue
			}

			thresholdPercentGauge.With(prometheus.Labels{
				"chain":    consumer.Name,
				"chain_id": consumer.ChainID,
				"provider": chain.Name,
			}).Set(float64(consumerInfo.TopN) / 100)

			minStakeGauge.With(prometheus.Labels{
				"chain":    consumer.Name,
				"chain_id": consumer.ChainID,
				"provider": chain.Name,
			}).Set(float64(consumerInfo.MinPowerInTopN.Int64()))
		}
	}

	return []prometheus.Collector{thresholdPercentGauge, minStakeGauge}
}
