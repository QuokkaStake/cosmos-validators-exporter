package generators

import (
	"github.com/prometheus/client_golang/prometheus"
	"main/pkg/config"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
)

type SupplyGenerator struct {
	Chains []*config.Chain
}

func NewSupplyGenerator(chains []*config.Chain) *SupplyGenerator {
	return &SupplyGenerator{Chains: chains}
}

func (g *SupplyGenerator) Generate(state fetchersPkg.State) []prometheus.Collector {
	data, ok := fetchersPkg.StateGet[fetchersPkg.SupplyData](state, constants.FetcherNameSupply)
	if !ok {
		return []prometheus.Collector{}
	}

	supplyGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "supply",
			Help: "Total chain supply",
		},
		[]string{"chain", "denom"},
	)

	for _, chain := range g.Chains {
		for _, consumer := range chain.ConsumerChains {
			consumerSupplies, ok := data.Supplies[consumer.Name]
			if !ok {
				continue
			}

			for _, supply := range consumerSupplies {
				amountConverted := consumer.Denoms.Convert(&supply)
				if amountConverted == nil {
					continue
				}

				supplyGauge.With(prometheus.Labels{
					"chain": consumer.Name,
					"denom": amountConverted.Denom,
				}).Set(amountConverted.Amount)
			}
		}

		chainSupplies, ok := data.Supplies[chain.Name]
		if !ok {
			continue
		}

		for _, balance := range chainSupplies {
			amountConverted := chain.Denoms.Convert(&balance)
			if amountConverted == nil {
				continue
			}

			supplyGauge.With(prometheus.Labels{
				"chain": chain.Name,
				"denom": amountConverted.Denom,
			}).Set(amountConverted.Amount)
		}
	}

	return []prometheus.Collector{supplyGauge}
}
