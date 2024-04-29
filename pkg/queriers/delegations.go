package queriers

import (
	"context"
	"main/pkg/config"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"
	"sync"

	"go.opentelemetry.io/otel/trace"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type DelegationsQuerier struct {
	Logger zerolog.Logger
	Config *config.Config
	Tracer trace.Tracer
}

func NewDelegationsQuerier(logger *zerolog.Logger, config *config.Config, tracer trace.Tracer) *DelegationsQuerier {
	return &DelegationsQuerier{
		Logger: logger.With().Str("component", "delegations_querier").Logger(),
		Config: config,
		Tracer: tracer,
	}
}

func (q *DelegationsQuerier) GetMetrics(ctx context.Context) ([]prometheus.Collector, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

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
		rpc := tendermint.NewRPC(chain, q.Config.Timeout, q.Logger, q.Tracer)

		for _, validator := range chain.Validators {
			wg.Add(1)
			go func(validator string, rpc *tendermint.RPC, chain config.Chain) {
				defer wg.Done()
				delegatorsResponse, query, err := rpc.GetDelegationsCount(validator, ctx)

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
						Msg("Error querying validator delegators count")
					return
				}

				if delegatorsResponse == nil {
					return
				}

				delegationsCountGauge.With(prometheus.Labels{
					"chain":   chain.Name,
					"address": validator,
				}).Set(float64(utils.StrToInt64(delegatorsResponse.Pagination.Total)))
			}(validator.Address, rpc, chain)
		}
	}

	wg.Wait()

	return []prometheus.Collector{delegationsCountGauge}, queryInfos
}

func (q *DelegationsQuerier) Name() string {
	return "delegations-querier"
}
