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

type ChainInfoQuerier struct {
	Logger zerolog.Logger
	Config *config.Config
	Tracer trace.Tracer
}

func NewChainInfoQuerier(
	logger *zerolog.Logger,
	config *config.Config,
	tracer trace.Tracer,
) *ChainInfoQuerier {
	return &ChainInfoQuerier{
		Logger: logger.With().Str("component", "chain_info_querier").Logger(),
		Config: config,
		Tracer: tracer,
	}
}

func (q *ChainInfoQuerier) GetMetrics(
	ctx context.Context,
) ([]prometheus.Collector, []*types.QueryInfo) {
	isConsumerGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_is_consumer",
			Help: "Whether the chain is consumer (1 if yes, 0 if no)",
		},
		[]string{"chain"},
	)

	softOptOutThresholdGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_soft_opt_out_threshold",
			Help: "Consumer chain's soft opt-out threshold",
		},
		[]string{"chain"},
	)

	queryInfos := make([]*types.QueryInfo, 0)

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Config.Chains {
		isConsumerGauge.With(prometheus.Labels{
			"chain": chain.Name,
		}).Set(utils.BoolToFloat64(chain.IsConsumer()))

		if chain.IsConsumer() {
			wg.Add(1)

			go func(chain config.Chain) {
				defer wg.Done()

				rpc := tendermint.NewRPC(chain, q.Config.Timeout, q.Logger, q.Tracer)

				threshold, query, err := rpc.GetConsumerSoftOutOutThreshold(ctx)

				mutex.Lock()
				defer mutex.Unlock()

				if query != nil {
					queryInfos = append(queryInfos, query)
				}

				if err != nil {
					q.Logger.Error().
						Err(err).
						Str("chain", chain.Name).
						Msg("Error querying soft opt-out threshold")
					return
				}

				softOptOutThresholdGauge.With(prometheus.Labels{
					"chain": chain.Name,
				}).Set(threshold)
			}(chain)
		}
	}

	wg.Wait()

	return []prometheus.Collector{isConsumerGauge, softOptOutThresholdGauge}, queryInfos
}

func (q *ChainInfoQuerier) Name() string {
	return "chain-info-querier"
}
