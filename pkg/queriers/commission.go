package queriers

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"main/pkg/config"
	"main/pkg/tendermint"
	"main/pkg/types"
	"sync"
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

func (q *CommissionQuerier) GetMetrics() ([]prometheus.Collector, []types.QueryInfo) {
	var collectors []prometheus.Collector
	var queryInfos []types.QueryInfo

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Config.Chains {
		rpc := tendermint.NewRPC(chain, q.Config.Timeout, q.Logger)

		for _, validator := range chain.Validators {
			wg.Add(1)
			go func(validator string, rpc *tendermint.RPC) {
				defer wg.Done()
				commission, commissionQuery, commissionQueryError := rpc.GetValidatorCommission(validator)

				mutex.Lock()
				defer mutex.Unlock()

				queryInfos = append(queryInfos, commissionQuery)

				if commissionQueryError != nil {
					q.Logger.Error().
						Err(commissionQueryError).
						Str("chain", chain.Name).
						Str("address", validator).
						Msg("Error querying validator commission")
					return
				}

				commissionUnclaimedTokens := prometheus.NewGaugeVec(
					prometheus.GaugeOpts{
						Name: "cosmos_validators_exporter_unclaimed_commission",
						Help: "Validator's unclaimed commission (in tokens)",
					},
					[]string{"chain", "address", "denom"},
				)

				for _, balance := range commission {
					commissionUnclaimedTokens.With(prometheus.Labels{
						"chain":   chain.Name,
						"address": validator,
						"denom":   balance.Denom,
					}).Set(balance.Amount)
				}
			}(validator, rpc)
		}
	}

	wg.Wait()

	return collectors, queryInfos
}
