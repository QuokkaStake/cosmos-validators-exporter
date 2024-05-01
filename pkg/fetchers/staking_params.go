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
	Tracer trace.Tracer
}

type StakingParamsData struct {
	Params map[string]*types.StakingParamsResponse
}

func NewStakingParamsFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	tracer trace.Tracer,
) *StakingParamsFetcher {
	return &StakingParamsFetcher{
		Logger: logger.With().Str("component", "staking_params_fetcher").Logger(),
		Config: config,
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
		rpc := tendermint.NewRPC(chain, q.Config.Timeout, q.Logger, q.Tracer)

		wg.Add(1)

		go func(chain config.Chain, rpc *tendermint.RPC) {
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
					Msg("Error querying slashing params")
				return
			}

			if params != nil {
				allParams[chain.Name] = params
			}
		}(chain, rpc)
	}

	wg.Wait()

	return StakingParamsData{Params: allParams}, queryInfos
}

func (q *StakingParamsFetcher) Name() constants.FetcherName {
	return constants.FetcherNameStakingParams
}
