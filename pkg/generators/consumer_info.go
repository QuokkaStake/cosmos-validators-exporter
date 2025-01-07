package generators

import (
	"fmt"
	"main/pkg/config"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/utils"

	"github.com/prometheus/client_golang/prometheus"
)

type ConsumerInfoGenerator struct {
	Chains []*config.Chain
}

func NewConsumerInfoGenerator(chains []*config.Chain) *ConsumerInfoGenerator {
	return &ConsumerInfoGenerator{Chains: chains}
}

func (g *ConsumerInfoGenerator) Generate(state statePkg.State) []prometheus.Collector {
	consumerInfos, ok := statePkg.StateGet[fetchersPkg.ConsumerInfoData](state, constants.FetcherNameConsumerInfo)
	if !ok {
		return []prometheus.Collector{}
	}

	consumerInfoGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "consumer_info",
			Help: "Consumer chain info",
		},
		[]string{
			"chain_id",
			"consumer_id",
			"phase",
			"allow_inactive_vals",
			"provider",
		},
	)

	thresholdPercentGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "consumer_top_n",
			Help: "Top-N percent threshold for consumer chains.",
		},
		[]string{
			"consumer_id",
			"provider",
		},
	)

	minStakeGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "consumer_minimal_stake",
			Help: "Minimal stake to be required to sign blocks on this consumer chain",
		},
		[]string{
			"consumer_id",
			"provider",
		},
	)

	for _, chain := range g.Chains {
		consumersInfo, ok := consumerInfos.Info[chain.Name]
		if !ok {
			continue
		}

		for _, consumerInfo := range consumersInfo {
			consumerInfoGauge.With(prometheus.Labels{
				"chain_id":            consumerInfo.ChainID,
				"consumer_id":         consumerInfo.ConsumerID,
				"phase":               consumerInfo.Phase,
				"allow_inactive_vals": fmt.Sprintf("%.0f", utils.BoolToFloat64(consumerInfo.AllowInactiveVals)),
				"provider":            chain.Name,
			}).Set(1)

			thresholdPercentGauge.With(prometheus.Labels{
				"consumer_id": consumerInfo.ConsumerID,
				"provider":    chain.Name,
			}).Set(float64(consumerInfo.TopN) / 100)

			minStakeGauge.With(prometheus.Labels{
				"consumer_id": consumerInfo.ConsumerID,
				"provider":    chain.Name,
			}).Set(float64(consumerInfo.MinPowerInTopN.Int64()))
		}
	}

	return []prometheus.Collector{consumerInfoGauge, thresholdPercentGauge, minStakeGauge}
}
