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

type UnbondsQuerier struct {
	Logger zerolog.Logger
	Config *config.Config
}

func NewUnbondsQuerier(logger *zerolog.Logger, config *config.Config) *UnbondsQuerier {
	return &UnbondsQuerier{
		Logger: logger.With().Str("component", "unbonds_querier").Logger(),
		Config: config,
	}
}

func (q *UnbondsQuerier) GetMetrics() ([]prometheus.Collector, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	unbondsCountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_unbonds_count",
			Help: "Validator unbonds count",
		},
		[]string{"chain", "address"},
	)

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Config.Chains {
		rpc := tendermint.NewRPC(chain, q.Config.Timeout, q.Logger)

		for _, validator := range chain.Validators {
			wg.Add(1)
			go func(validator string, rpc *tendermint.RPC, chain config.Chain) {
				defer wg.Done()
				unbondsResponse, query, err := rpc.GetUnbondsCount(validator)

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

				unbondsCountGauge.With(prometheus.Labels{
					"chain":   chain.Name,
					"address": validator,
				}).Set(float64(utils.StrToInt64(unbondsResponse.Pagination.Total)))
			}(validator.Address, rpc, chain)
		}
	}

	wg.Wait()

	return []prometheus.Collector{unbondsCountGauge}, queryInfos
}
