package generators

import (
	"main/pkg/config"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type DenomCoefficientGenerator struct {
	Chains []config.Chain
}

func NewDenomCoefficientGenerator(chains []config.Chain) *DenomCoefficientGenerator {
	return &DenomCoefficientGenerator{Chains: chains}
}

func (g *DenomCoefficientGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	denomCoefficientGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_denom_coefficient",
			Help: "Denom coefficient info",
		},
		[]string{"chain", "denom", "display_denom"},
	)

	baseDenomGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_base_denom",
			Help: "Base denom info",
		},
		[]string{"chain", "denom"},
	)

	for _, chain := range g.Chains {
		if chain.BaseDenom != "" {
			baseDenomGauge.With(prometheus.Labels{
				"chain": chain.Name,
				"denom": chain.BaseDenom,
			}).Set(float64(1))
		}

		for _, denom := range chain.Denoms {
			denomCoefficientGauge.With(prometheus.Labels{
				"chain":         chain.Name,
				"display_denom": denom.DisplayDenom,
				"denom":         denom.Denom,
			}).Set(float64(denom.DenomCoefficient))
		}
	}

	return []prometheus.Collector{denomCoefficientGauge, baseDenomGauge}
}
