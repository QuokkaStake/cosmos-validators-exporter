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
	dataRaw, ok := state.Get(constants.FetcherNameSigningInfo)
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

	data, _ := dataRaw.(fetchersPkg.SigningInfoData)

	for chain, commissions := range data.SigningInfos {
		for validator, signingInfo := range commissions {
			if signingInfo.MissedBlocksCounter >= 0 {
				missedBlocksGauge.With(prometheus.Labels{
					"chain":   chain,
					"address": validator,
				}).Set(float64(signingInfo.MissedBlocksCounter))
			}
		}
	}

	return []prometheus.Collector{missedBlocksGauge}
}
