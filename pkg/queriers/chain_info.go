package queriers

import (
	"context"
	"main/pkg/config"
	"main/pkg/types"
	"main/pkg/utils"

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

	queryInfos := make([]*types.QueryInfo, 0)

	for _, chain := range q.Config.Chains {
		isConsumerGauge.With(prometheus.Labels{
			"chain": chain.Name,
		}).Set(utils.BoolToFloat64(chain.IsConsumer()))
	}

	return []prometheus.Collector{isConsumerGauge}, queryInfos
}

func (q *ChainInfoQuerier) Name() string {
	return "chain-info-querier"
}
