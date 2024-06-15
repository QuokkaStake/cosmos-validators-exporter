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

type SlashingParamsFetcher struct {
	Logger zerolog.Logger
	Config *config.Config
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer
}

type SlashingParamsData struct {
	Params map[string]*types.SlashingParamsResponse
}

func NewSlashingParamsFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	rpcs map[string]*tendermint.RPCWithConsumers,
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

	allParams := map[string]*types.SlashingParamsResponse{}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	processChain := func(
		chainName string,
		rpc *tendermint.RPC,
		mutex *sync.Mutex,
		wg *sync.WaitGroup,
		allParams map[string]*types.SlashingParamsResponse,
	) {
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
				Str("chain", chainName).
				Msg("Error querying slashing params")
			return
		}

		if params != nil {
			allParams[chainName] = params
		}
	}

	for _, chain := range q.Config.Chains {
		rpc, _ := q.RPCs[chain.Name]

		wg.Add(1 + len(chain.ConsumerChains))

		go processChain(chain.Name, rpc.RPC, &mutex, &wg, allParams)

		for consumerIndex, consumerChain := range chain.ConsumerChains {
			consumerRPC := rpc.Consumers[consumerIndex]
			go processChain(consumerChain.Name, consumerRPC, &mutex, &wg, allParams)
		}
	}

	wg.Wait()

	return SlashingParamsData{Params: allParams}, queryInfos
}

func (q *SlashingParamsFetcher) Name() constants.FetcherName {
	return constants.FetcherNameSlashingParams
}
