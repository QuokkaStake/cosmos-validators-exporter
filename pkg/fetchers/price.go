package fetchers

import (
	"context"
	"main/pkg/config"
	"main/pkg/constants"
	coingeckoPkg "main/pkg/price_fetchers/coingecko"
	"main/pkg/types"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type PriceFetcher struct {
	Logger    zerolog.Logger
	Config    *config.Config
	Tracer    trace.Tracer
	Coingecko *coingeckoPkg.Coingecko

	CurrenciesRatesToChains map[string]map[string]float64
}

type PriceData struct {
	Prices map[string]map[string]float64
}

func NewPriceFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	tracer trace.Tracer,
	coingecko *coingeckoPkg.Coingecko,
) *PriceFetcher {
	return &PriceFetcher{
		Logger:    logger.With().Str("component", "price_fetcher").Logger(),
		Config:    config,
		Tracer:    tracer,
		Coingecko: coingecko,
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

	q.CurrenciesRatesToChains = map[string]map[string]float64{}

	for _, chain := range q.Config.Chains {
		q.CurrenciesRatesToChains[chain.Name] = make(map[string]float64)
		q.ProcessDenoms(chain.Name, chain.Denoms, currenciesRates)

		for _, consumer := range chain.ConsumerChains {
			q.CurrenciesRatesToChains[consumer.Name] = make(map[string]float64)
			q.ProcessDenoms(consumer.Name, consumer.Denoms, currenciesRates)
		}
	}

	return PriceData{Prices: q.CurrenciesRatesToChains}, queries
}

func (q *PriceFetcher) ProcessDenoms(chainName string, denoms config.DenomInfos, currenciesRates map[string]float64) {
	for _, denom := range denoms {
		// using coingecko response
		if rate, ok := currenciesRates[denom.CoingeckoCurrency]; ok {
			q.CurrenciesRatesToChains[chainName][denom.DisplayDenom] = rate
			continue
		}
	}
}

func (q *PriceFetcher) Name() constants.FetcherName {
	return constants.FetcherNamePrice
}
