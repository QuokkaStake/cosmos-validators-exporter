package generators

import (
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type PriceGenerator struct {
}

func NewPriceGenerator() *PriceGenerator {
	return &PriceGenerator{}
}

func (g *PriceGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	dataRaw, ok := state.Get(constants.FetcherNamePrice)
	if !ok {
		return []prometheus.Collector{}
	}

	tokenPriceGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "price",
			Help: "Price of 1 token in display denom in USD",
		},
		[]string{"chain", "denom"},
	)

	data, _ := dataRaw.(fetchersPkg.PriceData)

	for chainName, chainPrices := range data.Prices {
		for denom, price := range chainPrices {
			tokenPriceGauge.With(prometheus.Labels{
				"chain": chainName,
				"denom": denom,
			}).Set(price)
		}
	}

	return []prometheus.Collector{tokenPriceGauge}
}
