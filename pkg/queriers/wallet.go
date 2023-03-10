package queriers

import (
	"main/pkg/config"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type WalletQuerier struct {
	Logger zerolog.Logger
	Config *config.Config
}

func NewWalletQuerier(logger *zerolog.Logger, config *config.Config) *WalletQuerier {
	return &WalletQuerier{
		Logger: logger.With().Str("component", "wallet_querier").Logger(),
		Config: config,
	}
}

func (q *WalletQuerier) GetMetrics() ([]prometheus.Collector, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	var wg sync.WaitGroup
	var mutex sync.Mutex

	walletBalanceTokens := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_wallet_balance",
			Help: "Validator's wallet balance (in tokens)",
		},
		[]string{"chain", "address", "denom"},
	)

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

				balances, query, err := rpc.GetWalletBalance(wallet)

				mutex.Lock()
				defer mutex.Unlock()

				queryInfos = append(queryInfos, query)

				if err != nil {
					q.Logger.Error().
						Err(err).
						Str("chain", chain.Name).
						Str("address", validator).
						Msg("Error querying for validator wallet balance")
					return
				}

				for _, balance := range balances {
					walletBalanceTokens.With(prometheus.Labels{
						"chain":   chain.Name,
						"address": validator,
						"denom":   balance.Denom,
					}).Set(balance.Amount)
				}
			}(validator, rpc, chain)
		}
	}

	wg.Wait()

	return []prometheus.Collector{walletBalanceTokens}, queryInfos
}
