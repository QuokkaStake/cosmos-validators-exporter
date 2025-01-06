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

type SupplyFetcher struct {
	Logger zerolog.Logger
	Chains []*config.Chain
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer

	wg    sync.WaitGroup
	mutex sync.Mutex

	queryInfos  []*types.QueryInfo
	allSupplies map[string][]types.Amount
}

type SupplyData struct {
	Supplies map[string][]types.Amount
}

func NewSupplyFetcher(
	logger *zerolog.Logger,
	chains []*config.Chain,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *SupplyFetcher {
	return &SupplyFetcher{
		Logger: logger.With().Str("component", "balance_fetcher").Logger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *SupplyFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}

func (q *SupplyFetcher) Fetch(
	ctx context.Context,
	data ...interface{},
) (interface{}, []*types.QueryInfo) {
	q.queryInfos = []*types.QueryInfo{}
	q.allSupplies = map[string][]types.Amount{}

	for _, chain := range q.Chains {
		q.wg.Add(1 + len(chain.ConsumerChains))

		rpc, _ := q.RPCs[chain.Name]

		go q.processChain(
			ctx,
			chain.Name,
			rpc.RPC,
		)

		for consumerIndex, consumerChain := range chain.ConsumerChains {
			consumerRPC := rpc.Consumers[consumerIndex]

			go q.processChain(
				ctx,
				consumerChain.Name,
				consumerRPC,
			)
		}
	}

	q.wg.Wait()

	return SupplyData{Supplies: q.allSupplies}, q.queryInfos
}

func (q *SupplyFetcher) Name() constants.FetcherName {
	return constants.FetcherNameSupply
}

func (q *SupplyFetcher) processChain(
	ctx context.Context,
	chainName string,
	rpc *tendermint.RPC,
) {
	defer q.wg.Done()

	supply, query, err := rpc.GetTotalSupply(ctx)

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if query != nil {
		q.queryInfos = append(q.queryInfos, query)
	}

	if err != nil {
		q.Logger.Error().
			Err(err).
			Str("chain", chainName).
			Msg("Error querying for chain supply")
		return
	}

	if supply == nil {
		return
	}

	q.allSupplies[chainName] = supply
}
