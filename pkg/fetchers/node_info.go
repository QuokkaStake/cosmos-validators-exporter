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
	Chains []*config.Chain
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer

	wg    sync.WaitGroup
	mutex sync.Mutex

	queryInfos   []*types.QueryInfo
	allNodeInfos map[string]*types.NodeInfoResponse
}

type NodeInfoData struct {
	NodeInfos map[string]*types.NodeInfoResponse
}

func NewNodeInfoFetcher(
	logger *zerolog.Logger,
	chains []*config.Chain,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *NodeInfoFetcher {
	return &NodeInfoFetcher{
		Logger: logger.With().Str("component", "node_info_fetcher").Logger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *NodeInfoFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}

func (q *NodeInfoFetcher) Fetch(
	ctx context.Context,
	data ...interface{},
) (interface{}, []*types.QueryInfo) {
	q.queryInfos = []*types.QueryInfo{}
	q.allNodeInfos = map[string]*types.NodeInfoResponse{}

	for _, chain := range q.Chains {
		rpc, _ := q.RPCs[chain.Name]

		q.wg.Add(1 + len(chain.ConsumerChains))

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

	return NodeInfoData{NodeInfos: q.allNodeInfos}, q.queryInfos
}

func (q *NodeInfoFetcher) Name() constants.FetcherName {
	return constants.FetcherNameNodeInfo
}

func (q *NodeInfoFetcher) processChain(
	ctx context.Context,
	chainName string,
	rpc *tendermint.RPC,
) {
	defer q.wg.Done()
	nodeInfo, query, err := rpc.GetNodeInfo(ctx)

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if query != nil {
		q.queryInfos = append(q.queryInfos, query)
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

	q.allNodeInfos[chainName] = nodeInfo
}
