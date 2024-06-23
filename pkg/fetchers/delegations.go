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

type DelegationsFetcher struct {
	Logger zerolog.Logger
	Chains []*config.Chain
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer
}

type DelegationsData struct {
	Delegations map[string]map[string]uint64
}

func NewDelegationsFetcher(
	logger *zerolog.Logger,
	chains []*config.Chain,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *DelegationsFetcher {
	return &DelegationsFetcher{
		Logger: logger.With().Str("component", "delegations_fetcher").Logger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *DelegationsFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	allDelegations := map[string]map[string]uint64{}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Chains {
		allDelegations[chain.Name] = map[string]uint64{}
		for _, consumerChain := range chain.ConsumerChains {
			allDelegations[consumerChain.Name] = map[string]uint64{}
		}
	}

	for _, chain := range q.Chains {
		rpc, _ := q.RPCs[chain.Name]

		for _, validator := range chain.Validators {
			wg.Add(1)
			go func(validator string, rpc *tendermint.RPC, chain *config.Chain) {
				defer wg.Done()
				delegatorsResponse, query, err := rpc.GetDelegationsCount(validator, ctx)

				mutex.Lock()
				defer mutex.Unlock()

				if query != nil {
					queryInfos = append(queryInfos, query)
				}

				if err != nil {
					q.Logger.Error().
						Err(err).
						Str("chain", chain.Name).
						Str("address", validator).
						Msg("Error querying validator delegators count")
					return
				}

				if delegatorsResponse == nil {
					return
				}

				allDelegations[chain.Name][validator] = delegatorsResponse.Pagination.Total

				// consumer chains do not have staking module, so no delegations, therefore
				// we do not calculate it here
			}(validator.Address, rpc.RPC, chain)
		}
	}

	wg.Wait()

	return DelegationsData{Delegations: allDelegations}, queryInfos
}

func (q *DelegationsFetcher) Name() constants.FetcherName {
	return constants.FetcherNameDelegations
}
