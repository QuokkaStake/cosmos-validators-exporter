package fetchers

import (
	"context"
	"main/pkg/config"
	"main/pkg/constants"
	coingeckoPkg "main/pkg/price_fetchers/coingecko"
	dexScreenerPkg "main/pkg/price_fetchers/dex_screener"
	"main/pkg/types"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type PriceFetcher struct {
	Logger      zerolog.Logger
	Config      *config.Config
	Tracer      trace.Tracer
	Coingecko   *coingeckoPkg.Coingecko
	DexScreener *dexScreenerPkg.DexScreener
}

type PriceData struct {
	Prices map[string]map[string]float64
}

func NewPriceFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	tracer trace.Tracer,
	coingecko *coingeckoPkg.Coingecko,
	dexScreener *dexScreenerPkg.DexScreener,
) *PriceFetcher {
	return &PriceFetcher{
		Logger:      logger.With().Str("component", "price_fetcher").Logger(),
		Config:      config,
		Tracer:      tracer,
		Coingecko:   coingecko,
		DexScreener: dexScreener,
	}
}

func (q *PriceFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
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

	return PriceData{Prices: currenciesRatesToChains}, queries
}

func (q *PriceFetcher) Name() constants.FetcherName {
	return constants.FetcherNamePrice
}
