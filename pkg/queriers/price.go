package queriers

import (
	"context"
	"main/pkg/config"
	coingeckoPkg "main/pkg/price_fetchers/coingecko"
	dexScreenerPkg "main/pkg/price_fetchers/dex_screener"
	"main/pkg/types"

	"go.opentelemetry.io/otel/trace"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type PriceQuerier struct {
	Logger      zerolog.Logger
	Config      *config.Config
	Coingecko   *coingeckoPkg.Coingecko
	DexScreener *dexScreenerPkg.DexScreener
	Tracer      trace.Tracer
}

func NewPriceQuerier(
	logger *zerolog.Logger,
	config *config.Config,
	tracer trace.Tracer,
	coingecko *coingeckoPkg.Coingecko,
	dexScreener *dexScreenerPkg.DexScreener,
) *PriceQuerier {
	return &PriceQuerier{
		Logger:      logger.With().Str("component", "price_querier").Logger(),
		Config:      config,
		Coingecko:   coingecko,
		DexScreener: dexScreener,
		Tracer:      tracer,
	}
}

func (q *PriceQuerier) GetMetrics(ctx context.Context) ([]prometheus.Collector, []*types.QueryInfo) {
	currenciesList := q.Config.GetCoingeckoCurrencies()

	var currenciesRates map[string]float64
	var currenciesQuery *types.QueryInfo

	var queries []*types.QueryInfo

	if len(currenciesList) > 0 {
		currenciesRates, currenciesQuery = q.Coingecko.FetchPrices(currenciesList, ctx)
	}

	if currenciesQuery != nil {
		queries = append(queries, currenciesQuery)
	}

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

	return []prometheus.Collector{tokenPriceGauge}, queries
}

func (q *PriceQuerier) Name() string {
	return "price-querier"
}
