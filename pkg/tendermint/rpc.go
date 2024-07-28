package tendermint

import (
	"context"
	"fmt"
	"main/pkg/config"
	"main/pkg/http"
	"main/pkg/types"
	"main/pkg/utils"
	"strconv"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type RPC struct {
	ChainName    string
	ChainHost    string
	ChainQueries config.Queries
	Client       *http.Client
	Timeout      int
	Logger       zerolog.Logger
	Tracer       trace.Tracer

	LastHeight map[string]int64
	Mutex      sync.Mutex
}

func NewRPC(
	chain config.ChainInfo,
	timeout int,
	logger zerolog.Logger,
	tracer trace.Tracer,
) *RPC {
	return &RPC{
		ChainName:    chain.GetName(),
		ChainHost:    chain.GetHost(),
		ChainQueries: chain.GetQueries(),
		Client:       http.NewClient(&logger, chain.GetName(), tracer),
		Timeout:      timeout,
		Logger: logger.With().
			Str("component", "rpc").
			Str("chain", chain.GetName()).
			Logger(),
		Tracer:     tracer,
		LastHeight: map[string]int64{},
	}
}

func (rpc *RPC) GetDelegationsCount(
	address string,
	ctx context.Context,
) (*types.PaginationResponse, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("delegations") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching validator delegations",
		trace.WithAttributes(attribute.String("address", address)),
	)
	defer span.End()

	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/delegations?pagination.count_total=true&pagination.limit=1",
		rpc.ChainHost,
		address,
	)

	var response *types.PaginationResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.PaginationResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetUnbondsCount(
	address string,
	ctx context.Context,
) (*types.PaginationResponse, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("unbonds") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching validator unbonds",
		trace.WithAttributes(attribute.String("address", address)),
	)
	defer span.End()

	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/unbonding_delegations?pagination.count_total=true&pagination.limit=1",
		rpc.ChainHost,
		address,
	)

	var response *types.PaginationResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.PaginationResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetSingleDelegation(
	validator, wallet string,
	ctx context.Context,
) (*types.Amount, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("self-delegation") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching single delegation",
		trace.WithAttributes(
			attribute.String("validator", validator),
			attribute.String("wallet", wallet),
		),
	)
	defer span.End()

	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/delegations/%s",
		rpc.ChainHost,
		validator,
		wallet,
	)

	var response types.SingleDelegationResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return &types.Amount{}, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.Amount{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	amount := response.DelegationResponse.Balance.ToAmount()
	return &amount, &info, nil
}

func (rpc *RPC) GetAllValidators(
	ctx context.Context,
) (*types.ValidatorsResponse, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("validators") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching validators list",
	)
	defer span.End()

	url := rpc.ChainHost + "/cosmos/staking/v1beta1/validators?pagination.count_total=true&pagination.limit=1000"

	var response *types.ValidatorsResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.ValidatorsResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetConsumerValidators(
	ctx context.Context,
	chainId string,
) (*types.ConsumerValidatorsResponse, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("consumer-validators") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching consumer validators list",
	)
	defer span.End()

	url := rpc.ChainHost + "/interchain_security/ccv/provider/consumer_validators/" + chainId

	var response *types.ConsumerValidatorsResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.ConsumerValidatorsResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetConsumerInfo(
	ctx context.Context,
) (*types.ConsumerInfoResponse, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("consumer-info") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching consumer info",
	)
	defer span.End()

	url := rpc.ChainHost + "/interchain_security/ccv/provider/consumer_chains"

	var response *types.ConsumerInfoResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.ConsumerInfoResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetValidatorCommission(
	address string,
	ctx context.Context,
) ([]types.Amount, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("commission") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching validator commission",
		trace.WithAttributes(attribute.String("address", address)),
	)
	defer span.End()

	url := fmt.Sprintf(
		"%s/cosmos/distribution/v1beta1/validators/%s/commission",
		rpc.ChainHost,
		address,
	)

	var response *types.CommissionResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return []types.Amount{}, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return []types.Amount{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return utils.Map(response.Commission.Commission, func(amount types.ResponseAmount) types.Amount {
		return amount.ToAmount()
	}), &info, nil
}

func (rpc *RPC) GetDelegatorRewards(
	validator, wallet string,
	ctx context.Context,
) ([]types.Amount, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("rewards") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching delegator rewards",
		trace.WithAttributes(
			attribute.String("validator", validator),
			attribute.String("wallet", wallet),
		),
	)
	defer span.End()

	url := fmt.Sprintf(
		"%s/cosmos/distribution/v1beta1/delegators/%s/rewards/%s",
		rpc.ChainHost,
		wallet,
		validator,
	)

	var response *types.RewardsResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return []types.Amount{}, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return []types.Amount{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return utils.Map(response.Rewards, func(amount types.ResponseAmount) types.Amount {
		return amount.ToAmount()
	}), &info, nil
}

func (rpc *RPC) GetWalletBalance(
	wallet string,
	ctx context.Context,
) ([]types.Amount, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("balance") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching wallet balance",
		trace.WithAttributes(attribute.String("wallet", wallet)),
	)
	defer span.End()

	url := fmt.Sprintf(
		"%s/cosmos/bank/v1beta1/balances/%s",
		rpc.ChainHost,
		wallet,
	)

	var response types.BalancesResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return []types.Amount{}, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return nil, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return utils.Map(response.Balances, func(amount types.ResponseAmount) types.Amount {
		return amount.ToAmount()
	}), &info, nil
}

func (rpc *RPC) GetConsumerAssignedKey(
	valcons string,
	chainID string,
	ctx context.Context,
) (*types.AssignedKeyResponse, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("assigned-key") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching validator assigned key",
		trace.WithAttributes(
			attribute.String("valcons", valcons),
			attribute.String("chain-id", chainID),
		),
	)
	defer span.End()

	url := fmt.Sprintf(
		"%s/interchain_security/ccv/provider/validator_consumer_addr?chain_id=%s&provider_address=%s",
		rpc.ChainHost,
		chainID,
		valcons,
	)

	var response *types.AssignedKeyResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.AssignedKeyResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetSigningInfo(
	valcons string,
	ctx context.Context,
) (*types.SigningInfoResponse, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("signing-info") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching validator signing info",
		trace.WithAttributes(attribute.String("valcons", valcons)),
	)
	defer span.End()

	url := fmt.Sprintf("%s/cosmos/slashing/v1beta1/signing_infos/%s", rpc.ChainHost, valcons)

	var response *types.SigningInfoResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.SigningInfoResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetSlashingParams(
	ctx context.Context,
) (*types.SlashingParamsResponse, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("slashing-params") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching slashing params",
	)
	defer span.End()

	url := rpc.ChainHost + "/cosmos/slashing/v1beta1/params"

	var response *types.SlashingParamsResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return nil, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetConsumerSoftOutOutThreshold(
	ctx context.Context,
) (float64, bool, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("params") {
		return 0, false, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching soft opt-out threshold params",
	)
	defer span.End()

	var response *types.ParamsResponse
	info, err := rpc.Get(
		rpc.ChainHost+"/cosmos/params/v1beta1/params?subspace=ccvconsumer&key=SoftOptOutThreshold",
		&response,
		childQuerierCtx,
	)
	if err != nil {
		return 0.0, false, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return 0, false, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	valueStripped := strings.ReplaceAll(response.Param.Value, "\"", "")
	value, err := strconv.ParseFloat(valueStripped, 64)
	if err != nil {
		info.Success = false
		return 0, false, &info, err
	}

	return value, true, &info, nil
}

func (rpc *RPC) GetStakingParams(
	ctx context.Context,
) (*types.StakingParamsResponse, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("staking-params") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching staking params",
	)
	defer span.End()

	url := rpc.ChainHost + "/cosmos/staking/v1beta1/params"

	var response *types.StakingParamsResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.StakingParamsResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetNodeInfo(
	ctx context.Context,
) (*types.NodeInfoResponse, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("node-info") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching node info",
	)
	defer span.End()

	url := rpc.ChainHost + "/cosmos/base/tendermint/v1beta1/node_info"

	var response *types.NodeInfoResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.NodeInfoResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetValidatorConsumerChains(
	ctx context.Context,
	valcons string,
) (*types.ValidatorConsumerChains, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("validator-consumer-chains") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching validator required consumer chains",
	)
	defer span.End()

	url := rpc.ChainHost + "/interchain_security/ccv/provider/consumer_chains_per_validator/" + valcons

	var response *types.ValidatorConsumerChains
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.ValidatorConsumerChains{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetConsumerCommission(
	ctx context.Context,
	valcons string,
	chainID string,
) (*types.ConsumerCommissionResponse, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("consumer-commission") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching validator consumer commission",
	)
	defer span.End()

	url := rpc.ChainHost + "/interchain_security/ccv/provider/consumer_commission_rate/" + chainID + "/" + valcons

	var response *types.ConsumerCommissionResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.ConsumerCommissionResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetInflation(ctx context.Context) (*types.InflationResponse, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("inflation") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching chain inflation",
	)
	defer span.End()

	url := rpc.ChainHost + "/cosmos/mint/v1beta1/inflation"

	var response *types.InflationResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.InflationResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetTotalSupply(ctx context.Context) ([]types.Amount, *types.QueryInfo, error) {
	if !rpc.ChainQueries.Enabled("supply") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching chain supply",
	)
	defer span.End()

	url := rpc.ChainHost + "/cosmos/bank/v1beta1/supply?pagination.limit=10000&pagination.offset=0"

	var response *types.SupplyResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return []types.Amount{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return utils.Map(response.Supply, func(amount types.ResponseAmount) types.Amount {
		return amount.ToAmount()
	}), &info, nil
}

func (rpc *RPC) Get(
	url string,
	target interface{},
	ctx context.Context,
) (types.QueryInfo, error) {
	rpc.Mutex.Lock()
	previousHeight, found := rpc.LastHeight[url]
	if !found {
		previousHeight = 0
	}
	rpc.Mutex.Unlock()

	info, header, err := rpc.Client.Get(
		url,
		target,
		types.HTTPPredicateCheckHeightAfter(previousHeight),
		ctx,
	)

	if err != nil {
		return info, err
	}

	height, _ := utils.GetBlockHeightFromHeader(header)

	rpc.Mutex.Lock()
	rpc.LastHeight[url] = height
	rpc.Mutex.Unlock()

	rpc.Logger.Trace().
		Str("url", url).
		Int64("height", height).
		Msg("Got response at height")

	return info, err
}
