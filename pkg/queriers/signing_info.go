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

type SigningInfoQuerier struct {
	Logger zerolog.Logger
	Config *config.Config
}

func NewSigningInfoQuerier(logger *zerolog.Logger, config *config.Config) *SigningInfoQuerier {
	return &SigningInfoQuerier{
		Logger: logger.With().Str("component", "rewards_querier").Logger(),
		Config: config,
	}
}

func (q *SigningInfoQuerier) GetMetrics() ([]prometheus.Collector, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	missedBlocksGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_missed_blocks",
			Help: "Validator's missed blocks",
		},
		[]string{"chain", "address"},
	)

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Config.Chains {
		rpc := tendermint.NewRPC(chain, q.Config.Timeout, q.Logger)

		for _, validator := range chain.Validators {
			wg.Add(1)
			go func(validator config.Validator, rpc *tendermint.RPC, chain config.Chain) {
				defer wg.Done()

				if validator.ConsensusAddress == "" {
					return
				}

				signingInfo, signingInfoQuery, err := rpc.GetSigningInfo(validator.ConsensusAddress)

				mutex.Lock()
				defer mutex.Unlock()

				queryInfos = append(queryInfos, signingInfoQuery)

				if err != nil {
					q.Logger.Error().
						Err(err).
						Str("chain", chain.Name).
						Str("address", validator.Address).
						Msg("Error getting validator signing info")
					return
				}

				missedBlocksCounter := utils.StrToInt64(signingInfo.ValSigningInfo.MissedBlocksCounter)
				if missedBlocksCounter >= 0 {
					missedBlocksGauge.With(prometheus.Labels{
						"chain":   chain.Name,
						"address": validator.Address,
					}).Set(float64(missedBlocksCounter))
				}
			}(validator, rpc, chain)
		}
	}

	wg.Wait()

	return []prometheus.Collector{
		missedBlocksGauge,
	}, queryInfos
}
