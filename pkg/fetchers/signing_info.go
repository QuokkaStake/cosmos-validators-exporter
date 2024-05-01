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

type SigningInfoFetcher struct {
	Logger zerolog.Logger
	Config *config.Config
	Tracer trace.Tracer
}

type SigningInfoData struct {
	SigningInfos map[string]map[string]*types.SigningInfoResponse
}

func NewSigningInfoFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	tracer trace.Tracer,
) *SigningInfoFetcher {
	return &SigningInfoFetcher{
		Logger: logger.With().Str("component", "signing_infos").Logger(),
		Config: config,
		Tracer: tracer,
	}
}

func (q *SigningInfoFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	allSigningInfos := map[string]map[string]*types.SigningInfoResponse{}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Config.Chains {
		mutex.Lock()
		allSigningInfos[chain.Name] = map[string]*types.SigningInfoResponse{}
		mutex.Unlock()

		rpc := tendermint.NewRPC(chain, q.Config.Timeout, q.Logger, q.Tracer)

		for _, validator := range chain.Validators {
			wg.Add(1)
			go func(validator config.Validator, rpc *tendermint.RPC, chain config.Chain) {
				defer wg.Done()

				if validator.ConsensusAddress == "" {
					return
				}

				signingInfo, signingInfoQuery, err := rpc.GetSigningInfo(validator.ConsensusAddress, ctx)

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

				allSigningInfos[chain.Name][validator.Address] = signingInfo
			}(validator, rpc, chain)
		}
	}

	wg.Wait()

	return SigningInfoData{SigningInfos: allSigningInfos}, queryInfos
}

func (q *SigningInfoFetcher) Name() constants.FetcherName {
	return constants.FetcherNameSigningInfo
}
