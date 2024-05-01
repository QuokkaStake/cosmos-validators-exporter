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

type BalanceFetcher struct {
	Logger zerolog.Logger
	Config *config.Config
	Tracer trace.Tracer
}

type BalanceData struct {
	Balances map[string]map[string][]types.Amount
}

func NewBalanceFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	tracer trace.Tracer,
) *BalanceFetcher {
	return &BalanceFetcher{
		Logger: logger.With().Str("component", "balance_fetcher").Logger(),
		Config: config,
		Tracer: tracer,
	}
}

func (q *BalanceFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	allBalances := map[string]map[string][]types.Amount{}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Config.Chains {
		mutex.Lock()
		allBalances[chain.Name] = map[string][]types.Amount{}
		mutex.Unlock()

		rpc := tendermint.NewRPC(chain, q.Config.Timeout, q.Logger, q.Tracer)

		for _, validator := range chain.Validators {
			wg.Add(1)
			go func(validator string, rpc *tendermint.RPC, chain config.Chain) {
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

				balances, query, err := rpc.GetWalletBalance(wallet, ctx)

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
						Msg("Error querying for validator wallet balance")
					return
				}

				if balances == nil {
					return
				}

				allBalances[chain.Name][validator] = balances
			}(validator.Address, rpc, chain)
		}
	}

	wg.Wait()

	return BalanceData{Balances: allBalances}, queryInfos
}

func (q *BalanceFetcher) Name() constants.FetcherName {
	return constants.FetcherNameBalance
}
