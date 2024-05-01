package fetchers

import (
	"context"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/tendermint"
	"main/pkg/types"
	"sync"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type SoftOptOutThresholdFetcher struct {
	Logger zerolog.Logger
	Config *config.Config
	RPCs   map[string]*tendermint.RPC
	Tracer trace.Tracer
}

type SoftOptOutThresholdData struct {
	Thresholds map[string]float64
}

func NewSoftOptOutThresholdFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	rpcs map[string]*tendermint.RPC,
	tracer trace.Tracer,
) *SoftOptOutThresholdFetcher {
	return &SoftOptOutThresholdFetcher{
		Logger: logger.With().Str("component", "soft_opt_out_threshold_fetcher").Logger(),
		Config: config,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *SoftOptOutThresholdFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	allThresholds := map[string]float64{}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Config.Chains {
		rpc, _ := q.RPCs[chain.Name]

		if !chain.IsConsumer() {
			continue
		}

		wg.Add(1)

		go func(chain config.Chain, rpc *tendermint.RPC) {
			defer wg.Done()

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

			allThresholds[chain.Name] = threshold
		}(chain, rpc)
	}

	wg.Wait()

	return SoftOptOutThresholdData{Thresholds: allThresholds}, queryInfos
}

func (q *SoftOptOutThresholdFetcher) Name() constants.FetcherName {
	return constants.FetcherNameSoftOptOutThreshold
}
