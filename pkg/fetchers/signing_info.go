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
	Chains []*config.Chain
	RPCs   map[string]*tendermint.RPCWithConsumers
	Tracer trace.Tracer

	wg    sync.WaitGroup
	mutex sync.Mutex

	queryInfos      []*types.QueryInfo
	allSigningInfos map[string]map[string]*types.SigningInfoResponse
}

type SigningInfoData struct {
	SigningInfos map[string]map[string]*types.SigningInfoResponse
}

func NewSigningInfoFetcher(
	logger *zerolog.Logger,
	chains []*config.Chain,
	rpcs map[string]*tendermint.RPCWithConsumers,
	tracer trace.Tracer,
) *SigningInfoFetcher {
	return &SigningInfoFetcher{
		Logger: logger.With().Str("component", "signing_infos").Logger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *SigningInfoFetcher) Fetch(
	ctx context.Context,
) (interface{}, []*types.QueryInfo) {
	q.queryInfos = []*types.QueryInfo{}
	q.allSigningInfos = map[string]map[string]*types.SigningInfoResponse{}

	for _, chain := range q.Chains {
		q.allSigningInfos[chain.Name] = map[string]*types.SigningInfoResponse{}
		for _, consumerChain := range chain.ConsumerChains {
			q.allSigningInfos[consumerChain.Name] = map[string]*types.SigningInfoResponse{}
		}
	}

	for _, chain := range q.Chains {
		rpc, _ := q.RPCs[chain.Name]

		for _, validator := range chain.Validators {
			q.wg.Add(1 + len(rpc.Consumers))

			go q.processProviderChain(
				ctx,
				validator,
				chain.Name,
				rpc.RPC,
			)

			for consumerIndex, consumerChain := range chain.ConsumerChains {
				consumerRPC := rpc.Consumers[consumerIndex]

				go q.processConsumerChain(ctx, validator, consumerRPC, rpc.RPC, consumerChain)
			}
		}
	}

	q.wg.Wait()

	return SigningInfoData{SigningInfos: q.allSigningInfos}, q.queryInfos
}

func (q *SigningInfoFetcher) Name() constants.FetcherName {
	return constants.FetcherNameSigningInfo
}

func (q *SigningInfoFetcher) fetchAndSetSigningInfo(
	ctx context.Context,
	valoper string,
	valcons string,
	chainName string,
	rpc *tendermint.RPC,
) {
	if valcons == "" {
		return
	}

	signingInfo, signingInfoQuery, err := rpc.GetSigningInfo(valcons, ctx)

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if signingInfoQuery != nil {
		q.queryInfos = append(q.queryInfos, signingInfoQuery)
	}

	if err != nil {
		q.Logger.Error().
			Err(err).
			Str("chain", chainName).
			Str("address", valoper).
			Msg("Error getting validator signing info")
		return
	}

	q.allSigningInfos[chainName][valoper] = signingInfo
}

func (q *SigningInfoFetcher) processProviderChain(
	ctx context.Context,
	validator config.Validator,
	chainName string,
	rpc *tendermint.RPC,
) {
	defer q.wg.Done()

	q.fetchAndSetSigningInfo(
		ctx,
		validator.Address,
		validator.ConsensusAddress,
		chainName,
		rpc,
	)
}

func (q *SigningInfoFetcher) processConsumerChain(
	ctx context.Context,
	validator config.Validator,
	rpc *tendermint.RPC,
	providerRPC *tendermint.RPC,
	chain *config.ConsumerChain,
) {
	defer q.wg.Done()

	if chain.BechConsensusPrefix == "" || chain.BechValidatorPrefix == "" {
		return
	}

	// 1. Fetching assigned key.
	assignedKey, queryInfo, err := providerRPC.GetConsumerAssignedKey(
		validator.ConsensusAddress,
		chain.ChainID,
		ctx,
	)

	q.mutex.Lock()
	if queryInfo != nil {
		q.queryInfos = append(q.queryInfos, queryInfo)
	}
	q.mutex.Unlock()

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
	q.fetchAndSetSigningInfo(
		ctx,
		valoper,
		valcons,
		chain.Name,
		rpc,
	)
}
