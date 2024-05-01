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

type DelegationsFetcher struct {
	Logger zerolog.Logger
	Config *config.Config
	RPCs   map[string]*tendermint.RPC
	Tracer trace.Tracer
}

type DelegationsData struct {
	Delegations map[string]map[string]int64
}

func NewDelegationsFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	rpcs map[string]*tendermint.RPC,
	tracer trace.Tracer,
) *DelegationsFetcher {
	return &DelegationsFetcher{
		Logger: logger.With().Str("component", "delegations_fetcher").Logger(),
		Config: config,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *DelegationsFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	allDelegations := map[string]map[string]int64{}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Config.Chains {
		mutex.Lock()
		allDelegations[chain.Name] = map[string]int64{}
		mutex.Unlock()

		rpc, _ := q.RPCs[chain.Name]

		for _, validator := range chain.Validators {
			wg.Add(1)
			go func(validator string, rpc *tendermint.RPC, chain config.Chain) {
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

				allDelegations[chain.Name][validator] = utils.StrToInt64(delegatorsResponse.Pagination.Total)
			}(validator.Address, rpc, chain)
		}
	}

	wg.Wait()

	return DelegationsData{Delegations: allDelegations}, queryInfos
}

func (q *DelegationsFetcher) Name() constants.FetcherName {
	return constants.FetcherNameDelegations
}
