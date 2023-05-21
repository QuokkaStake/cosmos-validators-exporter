package queriers

import (
	"main/pkg/config"
	"main/pkg/tendermint"
	"main/pkg/types"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type CommissionQuerier struct {
	Logger zerolog.Logger
	Config *config.Config
}

func NewCommissionQuerier(logger *zerolog.Logger, config *config.Config) *CommissionQuerier {
	return &CommissionQuerier{
		Logger: logger.With().Str("component", "commission_querier").Logger(),
		Config: config,
	}
}

func (q *CommissionQuerier) GetMetrics() ([]prometheus.Collector, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	var wg sync.WaitGroup
	var mutex sync.Mutex

	commissionUnclaimedTokens := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_unclaimed_commission",
			Help: "Validator's unclaimed commission (in tokens)",
		},
		[]string{"chain", "address", "denom"},
	)

	for _, chain := range q.Config.Chains {
		rpc := tendermint.NewRPC(chain, q.Config.Timeout, q.Logger)

		for _, validator := range chain.Validators {
			wg.Add(1)
			go func(validator string, rpc *tendermint.RPC, chain config.Chain) {
				defer wg.Done()
				commission, query, err := rpc.GetValidatorCommission(validator)

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

				for _, balance := range commission {
					commissionUnclaimedTokens.With(prometheus.Labels{
						"chain":   chain.Name,
						"address": validator,
						"denom":   balance.Denom,
					}).Set(balance.Amount)
				}
			}(validator.Address, rpc, chain)
		}
	}

	wg.Wait()

	return []prometheus.Collector{commissionUnclaimedTokens}, queryInfos
}
