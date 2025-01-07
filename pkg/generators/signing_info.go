package generators

import (
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type SigningInfoGenerator struct {
}

func NewSigningInfoGenerator() *SigningInfoGenerator {
	return &SigningInfoGenerator{}
}

func (g *SigningInfoGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	data, ok := statePkg.StateGet[fetchersPkg.SigningInfoData](state, constants.FetcherNameSigningInfo)
	if !ok {
		return []prometheus.Collector{}
	}

	missedBlocksGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "missed_blocks",
			Help: "Validator's missed blocks",
		},
		[]string{"chain", "address"},
	)

	for chain, commissions := range data.SigningInfos {
		for validator, signingInfo := range commissions {
			missedBlocksCounter := signingInfo.ValSigningInfo.MissedBlocksCounter.Int64()
			if missedBlocksCounter >= 0 {
				missedBlocksGauge.With(prometheus.Labels{
					"chain":   chain,
					"address": validator,
				}).Set(float64(missedBlocksCounter))
			}
		}
	}

	return []prometheus.Collector{missedBlocksGauge}
}
