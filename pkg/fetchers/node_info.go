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
	RPCs   map[string]*tendermint.RPC
	Tracer trace.Tracer
}

type NodeInfoData struct {
	NodeInfos map[string]*types.NodeInfoResponse
}

func NewNodeInfoFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	rpcs map[string]*tendermint.RPC,
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

	for _, chain := range q.Config.Chains {
		rpc, _ := q.RPCs[chain.Name]

		wg.Add(1)
		go func(rpc *tendermint.RPC, chain config.Chain) {
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
					Str("chain", chain.Name).
					Msg("Error querying node info")
				return
			}

			if nodeInfo == nil {
				return
			}

			allNodeInfos[chain.Name] = nodeInfo
		}(rpc, chain)
	}

	wg.Wait()

	return NodeInfoData{NodeInfos: allNodeInfos}, queryInfos
}

func (q *NodeInfoFetcher) Name() constants.FetcherName {
	return constants.FetcherNameNodeInfo
}
