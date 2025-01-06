package fetchers

import (
	"context"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/price_fetchers"
	coingeckoPkg "main/pkg/price_fetchers/coingecko"
	"main/pkg/types"
	"sync"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type PriceFetcher struct {
	Logger zerolog.Logger
	Config *config.Config
	Tracer trace.Tracer

	Fetchers map[constants.PriceFetcherName]price_fetchers.PriceFetcher
}

type PriceInfo struct {
	Source       constants.PriceFetcherName
	BaseCurrency string
	Value        float64
}

type PriceData struct {
	Prices map[string]map[string]PriceInfo
}

func NewPriceFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	tracer trace.Tracer,
) *PriceFetcher {
	fetchers := map[constants.PriceFetcherName]price_fetchers.PriceFetcher{
		constants.PriceFetcherNameCoingecko: coingeckoPkg.NewCoingecko(config, logger, tracer),
	}

	return &PriceFetcher{
		Logger:   logger.With().Str("component", "price_fetcher").Logger(),
		Config:   config,
		Tracer:   tracer,
		Fetchers: fetchers,
	}
}

func (q *PriceFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}

func (q *PriceFetcher) Fetch(
	ctx context.Context,
	data ...interface{},
) (interface{}, []*types.QueryInfo) {
	queries := []*types.QueryInfo{}
	denomsByPriceFetcher := map[constants.PriceFetcherName][]price_fetchers.ChainWithDenom{}

	for _, chain := range q.Config.Chains {
		for _, denom := range chain.Denoms {
			for _, priceFetcher := range denom.PriceFetchers() {
				denomsByPriceFetcher[priceFetcher] = append(denomsByPriceFetcher[priceFetcher], price_fetchers.ChainWithDenom{
					Chain:     chain.Name,
					DenomInfo: denom,
				})
			}
		}

		for _, consumer := range chain.ConsumerChains {
			for _, denom := range consumer.Denoms {
				for _, priceFetcher := range denom.PriceFetchers() {
					denomsByPriceFetcher[priceFetcher] = append(denomsByPriceFetcher[priceFetcher], price_fetchers.ChainWithDenom{
						Chain:     consumer.Name,
						DenomInfo: denom,
					})
				}
			}
		}
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex
	var denomsPrices = map[constants.PriceFetcherName][]price_fetchers.PriceInfo{}

	for priceFetcher, denoms := range denomsByPriceFetcher {
		wg.Add(1)

		go func(priceFetcher constants.PriceFetcherName, denoms []price_fetchers.ChainWithDenom) {
			defer wg.Done()

			priceFetcherDenoms, priceFetcherQuery := q.Fetchers[priceFetcher].FetchPrices(denoms, ctx)

			mutex.Lock()
			queries = append(queries, priceFetcherQuery)
			denomsPrices[priceFetcher] = priceFetcherDenoms
			mutex.Unlock()
		}(priceFetcher, denoms)
	}

	wg.Wait()

	prices := map[string]map[string]PriceInfo{}

	for priceFetcher, denomInfos := range denomsPrices {
		for _, denomInfo := range denomInfos {
			if _, ok := prices[denomInfo.Chain]; !ok {
				prices[denomInfo.Chain] = map[string]PriceInfo{}
			}

			prices[denomInfo.Chain][denomInfo.Denom] = PriceInfo{
				Source:       priceFetcher,
				BaseCurrency: denomInfo.BaseCurrency,
				Value:        denomInfo.Price,
			}
		}
	}

	return PriceData{Prices: prices}, queries
}

func (q *PriceFetcher) Name() constants.FetcherName {
	return constants.FetcherNamePrice
}
