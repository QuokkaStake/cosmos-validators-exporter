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

type ValidatorsFetcher struct {
	Logger zerolog.Logger
	Config *config.Config
	RPCs   map[string]*tendermint.RPC
	Tracer trace.Tracer
}

type ValidatorsData struct {
	Validators map[string]*types.ValidatorsResponse
}

func NewValidatorsFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	rpcs map[string]*tendermint.RPC,
	tracer trace.Tracer,
) *ValidatorsFetcher {
	return &ValidatorsFetcher{
		Logger: logger.With().Str("component", "validators_fetcher").Logger(),
		Config: config,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (f *ValidatorsFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	allValidators := map[string]*types.ValidatorsResponse{}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range f.Config.Chains {
		rpc, _ := f.RPCs[chain.Name]

		wg.Add(1)
		go func(rpc *tendermint.RPC, chain config.Chain) {
			defer wg.Done()

			allValidatorsList, queryInfo, err := rpc.GetAllValidators(ctx)

			mutex.Lock()
			defer mutex.Unlock()

			if queryInfo != nil {
				queryInfos = append(queryInfos, queryInfo)
			}

			if err != nil {
				f.Logger.Error().
					Err(err).
					Str("chain", chain.Name).
					Msg("Error querying all validators")
				return
			}

			allValidators[chain.Name] = allValidatorsList
		}(rpc, chain)
	}

	wg.Wait()

	return ValidatorsData{Validators: allValidators}, queryInfos
}

func (q *ValidatorsFetcher) Name() constants.FetcherName {
	return constants.FetcherNameValidators
}
