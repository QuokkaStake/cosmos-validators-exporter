package generators

import (
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type StakingParamsGenerator struct {
}

func NewStakingParamsGenerator() *StakingParamsGenerator {
	return &StakingParamsGenerator{}
}

func (g *StakingParamsGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	dataRaw, ok := state.Get(constants.FetcherNameStakingParams)
	if !ok {
		return []prometheus.Collector{}
	}

	activeSetSizeGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "active_set_size",
			Help: "Active set size",
		},
		[]string{"chain"},
	)

	data, _ := dataRaw.(fetchersPkg.StakingParamsData)

	for chain, params := range data.Params {
		maxValidators := int64(params.MaxValidators)
		if maxValidators >= 0 {
			activeSetSizeGauge.With(prometheus.Labels{
				"chain": chain,
			}).Set(float64(maxValidators))
		}
	}

	return []prometheus.Collector{activeSetSizeGauge}
}
