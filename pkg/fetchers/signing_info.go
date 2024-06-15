package fetchers

import (
	"context"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"
	"sync"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type SigningInfoFetcher struct {
	Logger zerolog.Logger
	Config *config.Config
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer
}

type SigningInfoData struct {
	SigningInfos map[string]map[string]*types.SigningInfoResponse
}

func NewSigningInfoFetcher(
	logger *zerolog.Logger,
	config *config.Config,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *SigningInfoFetcher {
	return &SigningInfoFetcher{
		Logger: logger.With().Str("component", "signing_infos").Logger(),
		Config: config,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *SigningInfoFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	allSigningInfos := map[string]map[string]*types.SigningInfoResponse{}

	for _, chain := range q.Config.Chains {
		allSigningInfos[chain.Name] = map[string]*types.SigningInfoResponse{}
		for _, consumerChain := range chain.ConsumerChains {
			allSigningInfos[consumerChain.Name] = map[string]*types.SigningInfoResponse{}
		}
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	fetchAndSetConsumerKey := func(
		valoper string,
		valcons string,
		chainName string,
		rpc *tendermint.RPC,
		mutex *sync.Mutex,
	) {
		if valcons == "" {
			return
		}

		signingInfo, signingInfoQuery, err := rpc.GetSigningInfo(valcons, ctx)

		mutex.Lock()
		defer mutex.Unlock()

		queryInfos = append(queryInfos, signingInfoQuery)

		if err != nil {
			q.Logger.Error().
				Err(err).
				Str("chain", chainName).
				Str("address", valoper).
				Msg("Error getting validator signing info")
			return
		}

		allSigningInfos[chainName][valoper] = signingInfo
	}

	for _, chain := range q.Config.Chains {
		rpc, _ := q.RPCs[chain.Name]

		for _, validator := range chain.Validators {
			wg.Add(1 + len(rpc.Consumers))

			go func(validator config.Validator, rpc *tendermint.RPC, chain *config.Chain) {
				defer wg.Done()

				fetchAndSetConsumerKey(
					validator.Address,
					validator.ConsensusAddress,
					chain.Name,
					rpc,
					&mutex,
				)
			}(validator, rpc.RPC, chain)

			for consumerIndex, consumerChain := range chain.ConsumerChains {
				consumerRPC := rpc.Consumers[consumerIndex]

				go func(validator config.Validator, rpc *tendermint.RPC, providerRPC *tendermint.RPC, chain *config.ConsumerChain) {
					defer wg.Done()

					if chain.BechConsensusPrefix == "" {
						return
					}

					// 1. Fetching assigned key.
					assignedKey, queryInfo, err := providerRPC.GetConsumerAssignedKey(
						validator.ConsensusAddress,
						chain.ChainID,
						ctx,
					)

					mutex.Lock()
					queryInfos = append(queryInfos, queryInfo)
					mutex.Unlock()

					if err != nil {
						q.Logger.Error().
							Err(err).
							Str("chain", chain.Name).
							Str("address", validator.Address).
							Msg("Error getting validator assigned key")
						return
					}

					valconsProvider := validator.ConsensusAddress
					if assignedKey != nil && assignedKey.ConsumerAddress != "" {
						valconsProvider = assignedKey.ConsumerAddress
					}

					// 2. Converting it to bech32 prefix of the consumer chain.
					valcons, err := utils.ChangeBech32Prefix(valconsProvider, chain.BechConsensusPrefix)
					if err != nil {
						q.Logger.Error().
							Err(err).
							Str("chain", chain.Name).
							Str("address", validator.Address).
							Msg("Error converting valcons prefix")
						return
					}

					// 3. Converting valoper address on a provider chain to the one on a consumer chain.
					valoper, err := utils.ChangeBech32Prefix(validator.Address, chain.BechValidatorPrefix)
					if err != nil {
						q.Logger.Error().
							Err(err).
							Str("chain", chain.Name).
							Str("address", validator.Address).
							Msg("Error converting valoper prefix")
						return
					}

					// 4. Querying the signing-info on consumer chain.
					fetchAndSetConsumerKey(
						valoper,
						valcons,
						chain.Name,
						rpc,
						&mutex,
					)
				}(validator, consumerRPC, rpc.RPC, consumerChain)
			}
		}
	}

	wg.Wait()

	return SigningInfoData{SigningInfos: allSigningInfos}, queryInfos
}

func (q *SigningInfoFetcher) Name() constants.FetcherName {
	return constants.FetcherNameSigningInfo
}
