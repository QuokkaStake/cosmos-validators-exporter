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
	Chains []*config.Chain
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer
}

type SoftOptOutThresholdData struct {
	Thresholds map[string]float64
}

func NewSoftOptOutThresholdFetcher(
	logger *zerolog.Logger,
	chains []*config.Chain,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *SoftOptOutThresholdFetcher {
	return &SoftOptOutThresholdFetcher{
		Logger: logger.With().Str("component", "soft_opt_out_threshold_fetcher").Logger(),
		Chains: chains,
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

	for _, chain := range q.Chains {
		rpc, _ := q.RPCs[chain.Name]

		for consumerIndex, consumerChain := range chain.ConsumerChains {
			consumerRPC := rpc.Consumers[consumerIndex]

			wg.Add(1)

			go func(chain *config.ConsumerChain, rpc *tendermint.RPC) {
				defer wg.Done()

				threshold, queried, query, err := rpc.GetConsumerSoftOutOutThreshold(ctx)

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

				if queried {
					allThresholds[chain.Name] = threshold
				}
			}(consumerChain, consumerRPC)
		}
	}

	wg.Wait()

	return SoftOptOutThresholdData{Thresholds: allThresholds}, queryInfos
}

func (q *SoftOptOutThresholdFetcher) Name() constants.FetcherName {
	return constants.FetcherNameSoftOptOutThreshold
}
