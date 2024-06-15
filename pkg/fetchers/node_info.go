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

type NodeInfoFetcher struct {
	Logger zerolog.Logger
	Config *config.Config
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer
}

type NodeInfoData struct {
	NodeInfos map[string]*types.NodeInfoResponse
}

func NewNodeInfoFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *NodeInfoFetcher {
	return &NodeInfoFetcher{
		Logger: logger.With().Str("component", "commission_fetcher").Logger(),
		Config: config,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *NodeInfoFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	allNodeInfos := map[string]*types.NodeInfoResponse{}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	processChain := func(
		chainName string,
		rpc *tendermint.RPC,
		mutex *sync.Mutex,
		wg *sync.WaitGroup,
		allNodeInfos map[string]*types.NodeInfoResponse,
	) {
		defer wg.Done()
		nodeInfo, query, err := rpc.GetNodeInfo(ctx)

		mutex.Lock()
		defer mutex.Unlock()

		if query != nil {
			queryInfos = append(queryInfos, query)
		}

		if err != nil {
			q.Logger.Error().
				Err(err).
				Str("chain", chainName).
				Msg("Error querying node info")
			return
		}

		if nodeInfo == nil {
			return
		}

		allNodeInfos[chainName] = nodeInfo
	}

	for _, chain := range q.Config.Chains {
		rpc, _ := q.RPCs[chain.Name]

		wg.Add(1 + len(chain.ConsumerChains))

		go processChain(
			chain.Name,
			rpc.RPC,
			&mutex,
			&wg,
			allNodeInfos,
		)

		for consumerIndex, consumerChain := range chain.ConsumerChains {
			consumerRPC := rpc.Consumers[consumerIndex]
			go processChain(
				consumerChain.Name,
				consumerRPC,
				&mutex,
				&wg,
				allNodeInfos,
			)
		}
	}

	wg.Wait()

	return NodeInfoData{NodeInfos: allNodeInfos}, queryInfos
}

func (q *NodeInfoFetcher) Name() constants.FetcherName {
	return constants.FetcherNameNodeInfo
}
