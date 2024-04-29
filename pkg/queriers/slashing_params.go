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

type SlashingParamsQuerier struct {
	Logger zerolog.Logger
	Config *config.Config
	Tracer trace.Tracer
}

func NewSlashingParamsQuerier(
	logger *zerolog.Logger,
	config *config.Config,
	tracer trace.Tracer,
) *SlashingParamsQuerier {
	return &SlashingParamsQuerier{
		Logger: logger.With().Str("component", "slashing_params_querier").Logger(),
		Config: config,
		Tracer: tracer,
	}
}

func (q *SlashingParamsQuerier) GetMetrics(ctx context.Context) ([]prometheus.Collector, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	var wg sync.WaitGroup
	var mutex sync.Mutex

	blocksWindowGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_missed_blocks_window",
			Help: "Missed blocks window in network",
		},
		[]string{"chain"},
	)

	for _, chain := range q.Config.Chains {
		rpc := tendermint.NewRPC(chain, q.Config.Timeout, q.Logger, q.Tracer)

		wg.Add(1)

		go func(chain config.Chain, rpc *tendermint.RPC) {
			defer wg.Done()

			params, query, err := rpc.GetSlashingParams(ctx)

			mutex.Lock()
			defer mutex.Unlock()

			if query != nil {
				queryInfos = append(queryInfos, query)
			}

			if err != nil {
				q.Logger.Error().
					Err(err).
					Str("chain", chain.Name).
					Msg("Error querying slashing params")
				return
			}

			if params == nil {
				return
			}

			if params.SlashingParams.SignedBlocksWindow == "" {
				q.Logger.Error().
					Str("chain", chain.Name).
					Msg("Malformed response when querying for slashing params")
				return
			}

			blocksWindowGauge.With(prometheus.Labels{
				"chain": chain.Name,
			}).Set(float64(utils.StrToInt64(params.SlashingParams.SignedBlocksWindow)))
		}(chain, rpc)
	}

	wg.Wait()

	return []prometheus.Collector{blocksWindowGauge}, queryInfos
}

func (q *SlashingParamsQuerier) Name() string {
	return "slashing-params-querier"
}
