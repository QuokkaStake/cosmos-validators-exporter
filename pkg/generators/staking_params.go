package generators

import (
	"github.com/prometheus/client_golang/prometheus"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
)

type StakingParamsGenerator struct {
}

func NewStakingParamsGenerator() *StakingParamsGenerator {
	return &StakingParamsGenerator{}
}

func (g *StakingParamsGenerator) Generate(state fetchersPkg.State) []prometheus.Collector {
	data, ok := fetchersPkg.StateGet[fetchersPkg.StakingParamsData](state, constants.FetcherNameStakingParams)
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

	for chain, params := range data.Params {
		maxValidators := int64(params.StakingParams.MaxValidators)
		if maxValidators >= 0 {
			activeSetSizeGauge.With(prometheus.Labels{
				"chain": chain,
			}).Set(float64(maxValidators))
		}
	}

	return []prometheus.Collector{activeSetSizeGauge}
}
