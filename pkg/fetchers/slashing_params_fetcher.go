package fetchers

import (
	"context"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/tendermint"
	"main/pkg/types"
	"sync"
)

type SlashingParamsFetcher struct {
	Logger zerolog.Logger
	Config *config.Config
	Tracer trace.Tracer
}

type SlashingParamsData struct {
	Params map[string]*types.SlashingParamsResponse
}

func NewSlashingParamsFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	tracer trace.Tracer,
) *SlashingParamsFetcher {
	return &SlashingParamsFetcher{
		Logger: logger.With().Str("component", "slashing_params_fetcher").Logger(),
		Config: config,
		Tracer: tracer,
	}
}

func (q *SlashingParamsFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	allParams := map[string]*types.SlashingParamsResponse{}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Config.Chains {
		rpc := tendermint.NewRPC(chain, q.Config.Timeout, q.Logger, q.Tracer)

		wg.Add(1)

		go func(chain config.Chain, rpc *tendermint.RPC) {
			defer wg.Done()

			params, query, err := rpc.GetSlashingParams(ctx)

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

	return SlashingParamsData{Params: allParams}, queryInfos
}

func (q *SlashingParamsFetcher) Name() constants.FetcherName {
	return "slashing-params-fetcher"
}
