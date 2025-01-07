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

type ConsumerCommissionFetcher struct {
	Logger zerolog.Logger
	Chains []*config.Chain
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer

	wg    sync.WaitGroup
	mutex sync.Mutex

	queryInfos []*types.QueryInfo
	data       map[string]map[string]*types.ConsumerCommissionResponse
}

type ConsumerCommissionData struct {
	Commissions map[string]map[string]*types.ConsumerCommissionResponse
}

func NewConsumerCommissionFetcher(
	logger *zerolog.Logger,
	chains []*config.Chain,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *ConsumerCommissionFetcher {
	return &ConsumerCommissionFetcher{
		Logger: logger.With().Str("component", "consumer_commission_fetcher").Logger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (f *ConsumerCommissionFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}
func (f *ConsumerCommissionFetcher) Fetch(
	ctx context.Context,
	data ...interface{},
) (interface{}, []*types.QueryInfo) {
	f.queryInfos = []*types.QueryInfo{}
	f.data = map[string]map[string]*types.ConsumerCommissionResponse{}

	for _, chain := range f.Chains {
		for _, consumer := range chain.ConsumerChains {
			f.data[consumer.Name] = map[string]*types.ConsumerCommissionResponse{}
		}
	}

	for _, chain := range f.Chains {
		for _, validator := range chain.Validators {
			if validator.ConsensusAddress == "" {
				continue
			}

			rpc, _ := f.RPCs[chain.Name]

			for _, consumerChain := range chain.ConsumerChains {
				f.wg.Add(1)
				go f.processChain(ctx, rpc.RPC, consumerChain, validator)
			}
		}
	}

	f.wg.Wait()

	return ConsumerCommissionData{Commissions: f.data}, f.queryInfos
}

func (f *ConsumerCommissionFetcher) Name() constants.FetcherName {
	return constants.FetcherNameConsumerCommission
}

func (f *ConsumerCommissionFetcher) processChain(
	ctx context.Context,
	rpc *tendermint.RPC,
	chain *config.ConsumerChain,
	validator config.Validator,
) {
	defer f.wg.Done()

	commission, queryInfo, err := rpc.GetConsumerCommission(ctx, validator.ConsensusAddress, chain.ConsumerID)

	f.mutex.Lock()
	defer f.mutex.Unlock()

	if queryInfo != nil {
		f.queryInfos = append(f.queryInfos, queryInfo)
	}

	if err != nil {
		f.Logger.Error().
			Err(err).
			Str("chain", chain.Name).
			Msg("Error querying consumer commission")
		return
	}

	if commission != nil {
		f.data[chain.Name][validator.Address] = commission
	}
}
