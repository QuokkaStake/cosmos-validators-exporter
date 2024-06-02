package fetchers

import (
	"context"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/tendermint"
	"main/pkg/types"
	"sync"

	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type SlashingParamsFetcher struct {
	Logger zerolog.Logger
	Config *config.Config
	RPCs   map[string]*tendermint.RPC
	Tracer trace.Tracer
}

type SlashingParamsData struct {
	Params map[string]*slashingTypes.Params
}

func NewSlashingParamsFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	rpcs map[string]*tendermint.RPC,
	tracer trace.Tracer,
) *SlashingParamsFetcher {
	return &SlashingParamsFetcher{
		Logger: logger.With().Str("component", "slashing_params_fetcher").Logger(),
		Config: config,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *SlashingParamsFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	allParams := map[string]*slashingTypes.Params{}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Config.Chains {
		rpc, _ := q.RPCs[chain.Name]

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
	return constants.FetcherNameSlashingParams
}
