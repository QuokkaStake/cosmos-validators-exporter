package queriers

import (
	"main/pkg/config"
	"main/pkg/types"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type DenomCoefficientsQuerier struct {
	Logger zerolog.Logger
	Config *config.Config
}

func NewDenomCoefficientsQuerier(
	logger *zerolog.Logger,
	config *config.Config,
) *DenomCoefficientsQuerier {
	return &DenomCoefficientsQuerier{
		Logger: logger.With().Str("component", "denom_coefficients_querier").Logger(),
		Config: config,
	}
}

func (q *DenomCoefficientsQuerier) GetMetrics() ([]prometheus.Collector, []types.QueryInfo) {
	denomCoefficientGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_denom_coefficient",
			Help: "Denom coefficient info",
		},
		[]string{"chain", "denom", "display_denom"},
	)

	for _, chain := range q.Config.Chains {
		denomCoefficientGauge.With(prometheus.Labels{
			"chain":         chain.Name,
			"display_denom": chain.Denom,
			"denom":         chain.BaseDenom,
		}).Set(float64(chain.DenomCoefficient))
	}

	return []prometheus.Collector{denomCoefficientGauge}, []types.QueryInfo{}
}
