package generators

import (
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/utils"

	"github.com/prometheus/client_golang/prometheus"
)

type ConsumerInfoGenerator struct {
}

func NewConsumerInfoGenerator() *ConsumerInfoGenerator {
	return &ConsumerInfoGenerator{}
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
		},
	)

	for chain, chainConsumersInfo := range consumerInfos.Info {
		for _, info := range chainConsumersInfo.Chains {
			thresholdPercentGauge.With(prometheus.Labels{
				"chain":    chain,
				"chain_id": info.ChainID,
			}).Set(float64(info.TopN) / 100)

			minStakeGauge.With(prometheus.Labels{
				"chain":    chain,
				"chain_id": info.ChainID,
			}).Set(utils.StrToFloat64(info.MinPowerInTopN))
		}
	}

	return []prometheus.Collector{thresholdPercentGauge, minStakeGauge}
}
