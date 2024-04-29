package queriers

import (
	"context"
	"main/pkg/config"
	"main/pkg/tendermint"
	"main/pkg/types"
	"sync"

	"go.opentelemetry.io/otel/trace"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type CommissionQuerier struct {
	Logger zerolog.Logger
	Config *config.Config
	Tracer trace.Tracer
}

func NewCommissionQuerier(logger *zerolog.Logger, config *config.Config, tracer trace.Tracer) *CommissionQuerier {
	return &CommissionQuerier{
		Logger: logger.With().Str("component", "commission_querier").Logger(),
		Config: config,
		Tracer: tracer,
	}
}

func (q *CommissionQuerier) GetMetrics(ctx context.Context) ([]prometheus.Collector, []*types.QueryInfo) {
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
		rpc := tendermint.NewRPC(chain, q.Config.Timeout, q.Logger, q.Tracer)

		for _, validator := range chain.Validators {
			wg.Add(1)
			go func(validator string, rpc *tendermint.RPC, chain config.Chain) {
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

func (q *CommissionQuerier) Name() string {
	return "commissions-querier"
}
