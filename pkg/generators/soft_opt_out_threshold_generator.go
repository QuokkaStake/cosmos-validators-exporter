package generators

import (
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type SoftOptOutThresholdGenerator struct {
}

func NewSoftOptOutThresholdGenerator() *SoftOptOutThresholdGenerator {
	return &SoftOptOutThresholdGenerator{}
}

func (g *SoftOptOutThresholdGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	dataRaw, ok := state.Get(constants.FetcherNameSoftOptOutThreshold)
	if !ok {
		return []prometheus.Collector{}
	}

	softOptOutThresholdGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_soft_opt_out_threshold",
			Help: "Consumer chain's soft opt-out threshold",
		},
		[]string{"chain"},
	)

	data, _ := dataRaw.(fetchersPkg.SoftOptOutThresholdData)

	for chain, threshold := range data.Thresholds {
		softOptOutThresholdGauge.With(prometheus.Labels{
			"chain": chain,
		}).Set(threshold)
	}

	return []prometheus.Collector{softOptOutThresholdGauge}
}
