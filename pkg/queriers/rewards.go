package queriers

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"main/pkg/config"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"
	"sync"
)

type RewardsQuerier struct {
	Logger zerolog.Logger
	Config *config.Config
}

func NewRewardsQuerier(logger *zerolog.Logger, config *config.Config) *RewardsQuerier {
	return &RewardsQuerier{
		Logger: logger.With().Str("component", "rewards_querier").Logger(),
		Config: config,
	}
}

func (q *RewardsQuerier) GetMetrics() ([]prometheus.Collector, []types.QueryInfo) {
	var queryInfos []types.QueryInfo

	selfDelegationRewardsTokens := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_self_delegation_rewards",
			Help: "Validator's self-delegation rewards (in tokens)",
		},
		[]string{"chain", "address", "denom"},
	)

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Config.Chains {
		rpc := tendermint.NewRPC(chain, q.Config.Timeout, q.Logger)

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

				balances, query, err := rpc.GetDelegatorRewards(validator, wallet)
				if err != nil {
					q.Logger.Error().
						Err(err).
						Str("chain", chain.Name).
						Str("address", validator).
						Msg("Error querying for validator self-delegation rewards")
					return
				}

				mutex.Lock()
				defer mutex.Unlock()

				queryInfos = append(queryInfos, query)

				if err != nil {
					q.Logger.Error().
						Err(err).
						Str("chain", chain.Name).
						Str("address", validator).
						Msg("Error querying validator commission")
					return
				}

				for _, balance := range balances {
					selfDelegationRewardsTokens.With(prometheus.Labels{
						"chain":   chain.Name,
						"address": validator,
						"denom":   balance.Denom,
					}).Set(balance.Amount)
				}
			}(validator, rpc, chain)
		}
	}

	wg.Wait()

	return []prometheus.Collector{selfDelegationRewardsTokens}, queryInfos
}
