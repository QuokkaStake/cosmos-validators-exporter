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

type ConsumerInfoFetcher struct {
	Logger zerolog.Logger
	Chains []*config.Chain
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer

	wg    sync.WaitGroup
	mutex sync.Mutex

	queryInfos []*types.QueryInfo
	allInfos   map[string]map[string]types.ConsumerChainInfo
}

type ConsumerInfoData struct {
	Info map[string]map[string]types.ConsumerChainInfo
}

func NewConsumerInfoFetcher(
	logger *zerolog.Logger,
	chains []*config.Chain,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *ConsumerInfoFetcher {
	return &ConsumerInfoFetcher{
		Logger: logger.With().Str("component", "consumer_info_fetcher").Logger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (f *ConsumerInfoFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	f.queryInfos = []*types.QueryInfo{}
	f.allInfos = map[string]map[string]types.ConsumerChainInfo{}

	f.wg.Add(len(f.Chains))
	for _, chain := range f.Chains {
		rpc, _ := f.RPCs[chain.Name]
		go f.processChain(ctx, rpc.RPC, chain)
	}

	f.wg.Wait()

	return ConsumerInfoData{Info: f.allInfos}, f.queryInfos
}

func (f *ConsumerInfoFetcher) Name() constants.FetcherName {
	return constants.FetcherNameConsumerInfo
}

func (f *ConsumerInfoFetcher) processChain(
	ctx context.Context,
	rpc *tendermint.RPC,
	chain *config.Chain,
) {
	defer f.wg.Done()

	if !chain.IsProvider.Bool {
		return
	}

	allInfosList, queryInfo, err := rpc.GetConsumerInfo(ctx)

	f.mutex.Lock()
	defer f.mutex.Unlock()

	if queryInfo != nil {
		f.queryInfos = append(f.queryInfos, queryInfo)
	}

	if err != nil {
		f.Logger.Error().
			Err(err).
			Str("chain", chain.Name).
			Msg("Error querying consumer validators")
		return
	}

	if allInfosList == nil {
		return
	}

	f.allInfos[chain.Name] = map[string]types.ConsumerChainInfo{}

	for _, consumerInfo := range allInfosList.Chains {
		f.allInfos[chain.Name][consumerInfo.ChainID] = consumerInfo
	}
}
