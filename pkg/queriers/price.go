package queriers

import (
	"main/pkg/config"
	coingeckoPkg "main/pkg/price_fetchers/coingecko"
	dexScreenerPkg "main/pkg/price_fetchers/dex_screener"
	"main/pkg/types"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type PriceQuerier struct {
	Logger      zerolog.Logger
	Config      *config.Config
	Coingecko   *coingeckoPkg.Coingecko
	DexScreener *dexScreenerPkg.DexScreener
}

func NewPriceQuerier(
	logger *zerolog.Logger,
	config *config.Config,
	coingecko *coingeckoPkg.Coingecko,
	dexScreener *dexScreenerPkg.DexScreener,
) *PriceQuerier {
	return &PriceQuerier{
		Logger:      logger.With().Str("component", "price_querier").Logger(),
		Config:      config,
		Coingecko:   coingecko,
		DexScreener: dexScreener,
	}
}

func (q *PriceQuerier) GetMetrics() ([]prometheus.Collector, []*types.QueryInfo) {
	currenciesList := q.Config.GetCoingeckoCurrencies()
	currenciesRates, query := q.Coingecko.FetchPrices(currenciesList)

	currenciesRatesToChains := map[string]map[string]float64{}
	for _, chain := range q.Config.Chains {
		currenciesRatesToChains[chain.Name] = make(map[string]float64)

		for _, denom := range chain.Denoms {
			// using coingecko response
			if rate, ok := currenciesRates[denom.CoingeckoCurrency]; ok {
				currenciesRatesToChains[chain.Name][denom.Denom] = rate
				continue
			}

			// using dexscreener response
			if denom.DexScreenerChainID != "" && denom.DexScreenerPair != "" {
				rate, err := q.DexScreener.GetCurrency(denom.DexScreenerChainID, denom.DexScreenerPair)
				if err == nil {
					currenciesRatesToChains[chain.Name][denom.Denom] = rate
				}
			}
		}
	}

	tokenPriceGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_price",
			Help: "Price of 1 token in display denom in USD",
		},
		[]string{"chain", "denom"},
	)

	for chainName, chainPrices := range currenciesRatesToChains {
		for denom, price := range chainPrices {
			tokenPriceGauge.With(prometheus.Labels{
				"chain": chainName,
				"denom": denom,
			}).Set(price)
		}
	}

	return []prometheus.Collector{tokenPriceGauge}, []*types.QueryInfo{&query}
}
