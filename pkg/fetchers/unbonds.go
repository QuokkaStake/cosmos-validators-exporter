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

type UnbondsFetcher struct {
	Logger zerolog.Logger
	Config *config.Config
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer
}

type UnbondsData struct {
	Unbonds map[string]map[string]int64
}

func NewUnbondsFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *UnbondsFetcher {
	return &UnbondsFetcher{
		Logger: logger.With().Str("component", "unbonds_fetcher").Logger(),
		Config: config,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *UnbondsFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	allUnbonds := map[string]map[string]int64{}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Config.Chains {
		mutex.Lock()
		allUnbonds[chain.Name] = map[string]int64{}
		mutex.Unlock()

		rpc, _ := q.RPCs[chain.Name]

		for _, validator := range chain.Validators {
			wg.Add(1)
			go func(validator string, rpc *tendermint.RPC, chain *config.Chain) {
				defer wg.Done()
				unbondsResponse, query, err := rpc.GetUnbondsCount(validator, ctx)

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
						Msg("Error querying validator unbonding delegations count")
					return
				}

				if unbondsResponse == nil {
					return
				}

				allUnbonds[chain.Name][validator] = utils.StrToInt64(unbondsResponse.Pagination.Total)

				// consumer chains do not have staking module, so no unbonds, therefore
				// we do not calculate it here
			}(validator.Address, rpc.RPC, chain)
		}
	}

	wg.Wait()

	return UnbondsData{Unbonds: allUnbonds}, queryInfos
}

func (q *UnbondsFetcher) Name() constants.FetcherName {
	return constants.FetcherNameUnbonds
}
