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

	wg    sync.WaitGroup
	mutex sync.Mutex

	queryInfos []*types.QueryInfo
	allParams  map[string]*types.SlashingParamsResponse
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
	q.queryInfos = []*types.QueryInfo{}
	q.allParams = map[string]*types.SlashingParamsResponse{}

	for _, chain := range q.Config.Chains {
		rpc, _ := q.RPCs[chain.Name]

		q.wg.Add(1 + len(chain.ConsumerChains))

		go q.processChain(ctx, chain.Name, rpc.RPC)

		for consumerIndex, consumerChain := range chain.ConsumerChains {
			consumerRPC := rpc.Consumers[consumerIndex]
			go q.processChain(ctx, consumerChain.Name, consumerRPC)
		}
	}

	q.wg.Wait()

	return SlashingParamsData{Params: q.allParams}, q.queryInfos
}

func (q *SlashingParamsFetcher) Name() constants.FetcherName {
	return constants.FetcherNameSlashingParams
}

func (q *SlashingParamsFetcher) processChain(
	ctx context.Context,
	chainName string,
	rpc *tendermint.RPC,
) {
	defer q.wg.Done()

	params, query, err := rpc.GetSlashingParams(ctx)

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if query != nil {
		q.queryInfos = append(q.queryInfos, query)
	}

	if err != nil {
		q.Logger.Error().
			Err(err).
			Str("chain", chainName).
			Msg("Error querying slashing params")
		return
	}

	if params != nil {
		q.allParams[chainName] = params
	}
}
