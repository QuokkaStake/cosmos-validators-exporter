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

type DelegationsQuerier struct {
	Logger zerolog.Logger
	Config *config.Config
}

func NewDelegationsQuerier(logger *zerolog.Logger, config *config.Config) *DelegationsQuerier {
	return &DelegationsQuerier{
		Logger: logger.With().Str("component", "delegations_querier").Logger(),
		Config: config,
	}
}

func (q *DelegationsQuerier) GetMetrics() ([]prometheus.Collector, []types.QueryInfo) {
	var queryInfos []types.QueryInfo

	delegationsCountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_delegations_count",
			Help: "Validator delegations count",
		},
		[]string{"chain", "address"},
	)

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Config.Chains {
		rpc := tendermint.NewRPC(chain, q.Config.Timeout, q.Logger)

		for _, validator := range chain.Validators {
			wg.Add(1)
			go func(validator string, rpc *tendermint.RPC) {
				defer wg.Done()
				delegatorsResponse, query, err := rpc.GetDelegationsCount(validator)

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

				delegationsCountGauge.With(prometheus.Labels{
					"chain":   chain.Name,
					"address": validator,
				}).Set(float64(utils.StrToInt64(delegatorsResponse.Pagination.Total)))
			}(validator, rpc)
		}
	}

	wg.Wait()

	return []prometheus.Collector{delegationsCountGauge}, queryInfos
}