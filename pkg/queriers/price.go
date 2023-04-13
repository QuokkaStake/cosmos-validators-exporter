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

	currenciesRatesToChains := map[string]float64{}
	for _, chain := range q.Config.Chains {
		// using coingecko response
		if rate, ok := currenciesRates[chain.CoingeckoCurrency]; ok {
			currenciesRatesToChains[chain.Name] = rate
			continue
		}

		// using dexscreener response
		if chain.DexScreenerChainID != "" && chain.DexScreenerPair != "" {
			rate, err := q.DexScreener.GetCurrency(chain.DexScreenerChainID, chain.DexScreenerPair)
			if err == nil {
				currenciesRatesToChains[chain.Name] = rate
			}
		}
	}

	tokenPriceGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_price",
			Help: "Price of 1 token in display denom in USD",
		},
		[]string{"chain"},
	)

	for chain, price := range currenciesRatesToChains {
		tokenPriceGauge.With(prometheus.Labels{
			"chain": chain,
		}).Set(price)
	}

	return []prometheus.Collector{tokenPriceGauge}, []*types.QueryInfo{&query}
}
