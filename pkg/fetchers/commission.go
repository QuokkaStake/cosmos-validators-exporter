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

type CommissionFetcher struct {
	Logger zerolog.Logger
	Chains []*config.Chain
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer
}

type CommissionData struct {
	Commissions map[string]map[string][]types.Amount
}

func NewCommissionFetcher(
	logger *zerolog.Logger,
	chains []*config.Chain,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *CommissionFetcher {
	return &CommissionFetcher{
		Logger: logger.With().Str("component", "commission_fetcher").Logger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *CommissionFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}

func (q *CommissionFetcher) Fetch(
	ctx context.Context,
	data ...interface{},
) (interface{}, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	allCommissions := map[string]map[string][]types.Amount{}

	for _, chain := range q.Chains {
		allCommissions[chain.Name] = map[string][]types.Amount{}
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Chains {
		rpc, _ := q.RPCs[chain.Name]

		for _, validator := range chain.Validators {
			wg.Add(1)
			go func(validator string, rpc *tendermint.RPC, chain *config.Chain) {
				defer wg.Done()
				commission, query, err := rpc.GetValidatorCommission(validator, ctx)

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
						Msg("Error querying validator commission")
					return
				}

				if commission == nil {
					return
				}

				allCommissions[chain.Name][validator] = commission

				// consumers have no commission, so not counting it here
			}(validator.Address, rpc.RPC, chain)
		}
	}

	wg.Wait()

	return CommissionData{Commissions: allCommissions}, queryInfos
}

func (q *CommissionFetcher) Name() constants.FetcherName {
	return constants.FetcherNameCommission
}
