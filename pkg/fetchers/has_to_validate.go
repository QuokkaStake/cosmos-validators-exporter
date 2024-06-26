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

type ValidatorConsumersFetcher struct {
	Logger zerolog.Logger
	Chains []*config.Chain
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer

	wg    sync.WaitGroup
	mutex sync.Mutex

	queryInfos             []*types.QueryInfo
	allValidatorsConsumers map[string]map[string]map[string]bool
}

type ValidatorConsumersData struct {
	Infos map[string]map[string]map[string]bool
}

func NewValidatorConsumersFetcher(
	logger *zerolog.Logger,
	chains []*config.Chain,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *ValidatorConsumersFetcher {
	return &ValidatorConsumersFetcher{
		Logger: logger.With().Str("component", "node_info_fetcher").Logger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *ValidatorConsumersFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	q.queryInfos = []*types.QueryInfo{}
	q.allValidatorsConsumers = map[string]map[string]map[string]bool{}

	for _, chain := range q.Chains {
		if !chain.IsProvider.Bool {
			continue
		}

		q.allValidatorsConsumers[chain.Name] = map[string]map[string]bool{}

		rpc, _ := q.RPCs[chain.Name]

		q.wg.Add(len(chain.Validators))

		for _, validator := range chain.Validators {
			go q.processChain(
				ctx,
				chain.Name,
				rpc.RPC,
				validator,
			)
		}
	}

	q.wg.Wait()

	return ValidatorConsumersData{Infos: q.allValidatorsConsumers}, q.queryInfos
}

func (q *ValidatorConsumersFetcher) Name() constants.FetcherName {
	return constants.FetcherNameValidatorConsumers
}

func (q *ValidatorConsumersFetcher) processChain(
	ctx context.Context,
	chainName string,
	rpc *tendermint.RPC,
	validator config.Validator,
) {
	defer q.wg.Done()

	if validator.ConsensusAddress == "" {
		return
	}

	validatorConsumers, query, err := rpc.GetValidatorConsumerChains(ctx, validator.ConsensusAddress)

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

	if validatorConsumers == nil {
		return
	}

	q.allValidatorsConsumers[chainName][validator.Address] = map[string]bool{}
	for _, chainID := range validatorConsumers.ConsumerChainIds {
		q.allValidatorsConsumers[chainName][validator.Address][chainID] = true
	}
}
