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

type SlashingParamsQuerier struct {
	Logger zerolog.Logger
	Config *config.Config
}

func NewSlashingParamsQuerier(logger *zerolog.Logger, config *config.Config) *SlashingParamsQuerier {
	return &SlashingParamsQuerier{
		Logger: logger.With().Str("component", "slashing_params_querier").Logger(),
		Config: config,
	}
}

func (q *SlashingParamsQuerier) GetMetrics() ([]prometheus.Collector, []types.QueryInfo) {
	var queryInfos []types.QueryInfo

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
		rpc := tendermint.NewRPC(chain, q.Config.Timeout, q.Logger)

		wg.Add(1)

		go func(chain config.Chain, rpc *tendermint.RPC) {
			defer wg.Done()

			params, query, err := rpc.GetSlashingParams()

			mutex.Lock()
			defer mutex.Unlock()

			queryInfos = append(queryInfos, query)

			if err != nil {
				q.Logger.Error().
					Err(err).
					Str("chain", chain.Name).
					Msg("Error querying slashing params")
				return
			}

			if params == nil || params.SlashingParams.SignedBlocksWindow == "" {
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
