package fetchers

import (
	"context"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"
	"sync"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type SelfDelegationFetcher struct {
	Logger zerolog.Logger
	Chains []*config.Chain
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer
}

type SelfDelegationData struct {
	Delegations map[string]map[string]*types.Amount
}

func NewSelfDelegationFetcher(
	logger *zerolog.Logger,
	chains []*config.Chain,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *SelfDelegationFetcher {
	return &SelfDelegationFetcher{
		Logger: logger.With().Str("component", "self_delegation_fetcher").Logger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *SelfDelegationFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}

func (q *SelfDelegationFetcher) Fetch(
	ctx context.Context,
	data ...interface{},
) (interface{}, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	allSelfDelegations := map[string]map[string]*types.Amount{}

	for _, chain := range q.Chains {
		allSelfDelegations[chain.Name] = map[string]*types.Amount{}
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Chains {
		rpc, _ := q.RPCs[chain.Name]

		for _, validator := range chain.Validators {
			wg.Add(1)
			go func(validator string, rpc *tendermint.RPC, chain *config.Chain) {
				defer wg.Done()

				if chain.BechWalletPrefix == "" {
					return
				}

				wallet, err := utils.ChangeBech32Prefix(validator, chain.BechWalletPrefix)
				if err != nil {
					q.Logger.Error().
						Err(err).
						Str("chain", chain.Name).
						Str("address", validator).
						Msg("Error converting validator address")
					return
				}

				balance, query, err := rpc.GetSingleDelegation(validator, wallet, ctx)

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
						Msg("Error querying for validator self-delegation")
					return
				}

				if balance == nil {
					return
				}

				allSelfDelegations[chain.Name][validator] = balance

				// consumer chains do not have delegations, so not considering them here.
			}(validator.Address, rpc.RPC, chain)
		}
	}

	wg.Wait()

	return SelfDelegationData{Delegations: allSelfDelegations}, queryInfos
}

func (q *SelfDelegationFetcher) Name() constants.FetcherName {
	return constants.FetcherNameSelfDelegation
}
