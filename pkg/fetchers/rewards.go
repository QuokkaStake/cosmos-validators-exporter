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

type RewardsFetcher struct {
	Logger zerolog.Logger
	Config *config.Config
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer
}

type RewardsData struct {
	Rewards map[string]map[string][]types.Amount
}

func NewRewardsFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *RewardsFetcher {
	return &RewardsFetcher{
		Logger: logger.With().Str("component", "rewards_fetcher").Logger(),
		Config: config,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *RewardsFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	allRewards := map[string]map[string][]types.Amount{}
	for _, chain := range q.Config.Chains {
		allRewards[chain.Name] = map[string][]types.Amount{}
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Config.Chains {
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

				balances, query, err := rpc.GetDelegatorRewards(validator, wallet, ctx)

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
						Msg("Error querying for validator self-delegation rewards")
					return
				}

				if balances == nil {
					return
				}

				allRewards[chain.Name][validator] = balances

				// consumer chains have no rewards and/or staking module, so not counting it here
			}(validator.Address, rpc.RPC, chain)
		}
	}

	wg.Wait()

	return RewardsData{Rewards: allRewards}, queryInfos
}

func (q *RewardsFetcher) Name() constants.FetcherName {
	return constants.FetcherNameRewards
}
