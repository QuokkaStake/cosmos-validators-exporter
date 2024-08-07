package fetchers

import (
	"context"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/tendermint"
	"main/pkg/types"
	"sync"

	"cosmossdk.io/math"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type InflationFetcher struct {
	Logger zerolog.Logger
	Chains []*config.Chain
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer
}

type InflationData struct {
	Inflation map[string]math.LegacyDec
}

func NewInflationFetcher(
	logger *zerolog.Logger,
	chains []*config.Chain,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *InflationFetcher {
	return &InflationFetcher{
		Logger: logger.With().Str("component", "inflation_fetcher").Logger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *InflationFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	allInflation := map[string]math.LegacyDec{}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Chains {
		wg.Add(1)

		rpc, _ := q.RPCs[chain.Name]

		go func(rpc *tendermint.RPC, chain *config.Chain) {
			defer wg.Done()
			inflationResponse, query, err := rpc.GetInflation(ctx)

			mutex.Lock()
			defer mutex.Unlock()

			if query != nil {
				queryInfos = append(queryInfos, query)
			}

			if err != nil {
				q.Logger.Error().
					Err(err).
					Str("chain", chain.Name).
					Msg("Error querying inflation")
				return
			}

			if inflationResponse == nil {
				return
			}

			allInflation[chain.Name] = inflationResponse.Inflation

			// consumer chains do not have mint module, so no inflation, therefore
			// we do not calculate it here
		}(rpc.RPC, chain)
	}

	wg.Wait()

	return InflationData{Inflation: allInflation}, queryInfos
}

func (q *InflationFetcher) Name() constants.FetcherName {
	return constants.FetcherNameInflation
}
