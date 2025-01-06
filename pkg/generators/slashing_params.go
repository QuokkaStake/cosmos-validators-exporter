package generators

import (
	"github.com/prometheus/client_golang/prometheus"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
)

type SlashingParamsGenerator struct {
}

func NewSlashingParamsGenerator() *SlashingParamsGenerator {
	return &SlashingParamsGenerator{}
}

func (g *SlashingParamsGenerator) Generate(state fetchersPkg.State) []prometheus.Collector {
	data, ok := fetchersPkg.StateGet[fetchersPkg.SlashingParamsData](state, constants.FetcherNameSlashingParams)
	if !ok {
		return []prometheus.Collector{}
	}

	blocksWindowGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "missed_blocks_window",
			Help: "Missed blocks window in network",
		},
		[]string{"chain"},
	)

	for chain, params := range data.Params {
		blocksWindowGauge.With(prometheus.Labels{
			"chain": chain,
		}).Set(float64(params.SlashingParams.SignedBlocksWindow.Int64()))
	}

	return []prometheus.Collector{blocksWindowGauge}
}
