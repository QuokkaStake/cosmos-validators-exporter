package fetchers

import (
	"context"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/tendermint"
	"main/pkg/types"
	"sync"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type StakingParamsFetcher struct {
	Logger zerolog.Logger
	Config *config.Config
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer
}

type StakingParamsData struct {
	Params map[string]*types.StakingParamsResponse
}

func NewStakingParamsFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *StakingParamsFetcher {
	return &StakingParamsFetcher{
		Logger: logger.With().Str("component", "staking_params_fetcher").Logger(),
		Config: config,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *StakingParamsFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	allParams := map[string]*types.StakingParamsResponse{}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Config.Chains {
		rpc, _ := q.RPCs[chain.Name]

		wg.Add(1)

		// only fetching params for provider chains, as consumer chains
		// do not have the staking module, or it doesn't represent
		// the actual staking params (like on Stride)
		go func(chain *config.Chain, rpc *tendermint.RPC) {
			defer wg.Done()

			params, query, err := rpc.GetStakingParams(ctx)

			mutex.Lock()
			defer mutex.Unlock()

			if query != nil {
				queryInfos = append(queryInfos, query)
			}

			if err != nil {
				q.Logger.Error().
					Err(err).
					Str("chain", chain.Name).
					Msg("Error querying staking params")
				return
			}

			if params != nil {
				allParams[chain.Name] = params
			}
		}(chain, rpc.RPC)
	}

	wg.Wait()

	return StakingParamsData{Params: allParams}, queryInfos
}

func (q *StakingParamsFetcher) Name() constants.FetcherName {
	return constants.FetcherNameStakingParams
}
