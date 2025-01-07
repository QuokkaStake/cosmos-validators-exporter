package generators

import (
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type SlashingParamsGenerator struct {
}

func NewSlashingParamsGenerator() *SlashingParamsGenerator {
	return &SlashingParamsGenerator{}
}

func (g *SlashingParamsGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	data, ok := statePkg.StateGet[fetchersPkg.SlashingParamsData](state, constants.FetcherNameSlashingParams)
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
