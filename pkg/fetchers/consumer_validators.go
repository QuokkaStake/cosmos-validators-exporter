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

type ConsumerValidatorsFetcher struct {
	Logger zerolog.Logger
	Chains []*config.Chain
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer

	wg    sync.WaitGroup
	mutex sync.Mutex

	queryInfos    []*types.QueryInfo
	allValidators map[string]*types.ConsumerValidatorsResponse
}

type ConsumerValidatorsData struct {
	Validators map[string]*types.ConsumerValidatorsResponse
}

func NewConsumerValidatorsFetcher(
	logger *zerolog.Logger,
	chains []*config.Chain,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *ConsumerValidatorsFetcher {
	return &ConsumerValidatorsFetcher{
		Logger: logger.With().Str("component", "validators_fetcher").Logger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (f *ConsumerValidatorsFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	f.queryInfos = []*types.QueryInfo{}
	f.allValidators = map[string]*types.ConsumerValidatorsResponse{}

	for _, chain := range f.Chains {
		f.wg.Add(len(chain.ConsumerChains))

		rpc, _ := f.RPCs[chain.Name]

		for _, consumerChain := range chain.ConsumerChains {
			go f.processChain(ctx, rpc.RPC, consumerChain)
		}
	}

	f.wg.Wait()

	return ConsumerValidatorsData{Validators: f.allValidators}, f.queryInfos
}

func (f *ConsumerValidatorsFetcher) Name() constants.FetcherName {
	return constants.FetcherNameConsumerValidators
}

func (f *ConsumerValidatorsFetcher) processChain(
	ctx context.Context,
	rpc *tendermint.RPC,
	chain *config.ConsumerChain,
) {
	defer f.wg.Done()

	allValidatorsList, queryInfo, err := rpc.GetConsumerValidators(ctx, chain.ChainID)

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

	f.allValidators[chain.Name] = allValidatorsList
}
