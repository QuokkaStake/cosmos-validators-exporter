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
	Chain   config.Chain
	Client  *http.Client
	Timeout int
	Logger  zerolog.Logger
	Tracer  trace.Tracer

	LastHeight map[string]int64
	Mutex      sync.Mutex
}

func NewRPC(chain config.Chain, timeout int, logger zerolog.Logger, tracer trace.Tracer) *RPC {
	return &RPC{
		Chain:   chain,
		Client:  http.NewClient(&logger, chain.Name, tracer),
		Timeout: timeout,
		Logger: logger.With().
			Str("component", "rpc").
			Str("chain", chain.Name).
			Logger(),
		Tracer:     tracer,
		LastHeight: map[string]int64{},
	}
}

func (rpc *RPC) GetValidator(
	address string,
	ctx context.Context,
) (*types.ValidatorResponse, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("validator") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching validator",
		trace.WithAttributes(attribute.String("address", address)),
	)
	defer span.End()

	if rpc.Chain.IsConsumer() {
		return rpc.GetProviderValidator(address, childQuerierCtx)
	}

	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s",
		rpc.Chain.LCDEndpoint,
		address,
	)

	var response *types.ValidatorResponse
	info, err := rpc.Get(url, &response, ctx)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.ValidatorResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetProviderValidator(
	address string,
	ctx context.Context,
) (*types.ValidatorResponse, *types.QueryInfo, error) {
	if rpc.Chain.ProviderChainBechValidatorPrefix == "" {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching provider validator",
		trace.WithAttributes(attribute.String("address", address)),
	)
	defer span.End()

	providerAddress, err := utils.ChangeBech32Prefix(address, rpc.Chain.ProviderChainBechValidatorPrefix)
	if err != nil {
		return nil, nil, err
	}

	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s",
		rpc.Chain.ProviderChainLCD,
		providerAddress,
	)

	var response *types.ValidatorResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.ValidatorResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetDelegationsCount(
	address string,
	ctx context.Context,
) (*types.PaginationResponse, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("delegations") {
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
		rpc.Chain.LCDEndpoint,
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
	if !rpc.Chain.QueryEnabled("unbonds") {
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
		rpc.Chain.LCDEndpoint,
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
	if !rpc.Chain.QueryEnabled("self-delegation") {
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
		rpc.Chain.LCDEndpoint,
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
	if !rpc.Chain.QueryEnabled("validators") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching validators list",
	)
	defer span.End()

	host := rpc.Chain.LCDEndpoint
	if rpc.Chain.IsConsumer() {
		host = rpc.Chain.ProviderChainLCD
	}

	url := host + "/cosmos/staking/v1beta1/validators?pagination.count_total=true&pagination.limit=1000"

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

func (rpc *RPC) GetValidatorCommission(
	address string,
	ctx context.Context,
) ([]types.Amount, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("commission") {
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
		rpc.Chain.LCDEndpoint,
		address,
	)

	var response *types.CommissionResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return []types.Amount{}, &info, err
	}

	return utils.Map(response.Commission.Commission, func(amount types.ResponseAmount) types.Amount {
		return amount.ToAmount()
	}), &info, nil
}

func (rpc *RPC) GetDelegatorRewards(
	validator, wallet string,
	ctx context.Context,
) ([]types.Amount, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("rewards") {
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
		rpc.Chain.LCDEndpoint,
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
	if !rpc.Chain.QueryEnabled("balance") {
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
		rpc.Chain.LCDEndpoint,
		wallet,
	)

	var response types.BalancesResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return []types.Amount{}, &info, err
	}

	return utils.Map(response.Balances, func(amount types.ResponseAmount) types.Amount {
		return amount.ToAmount()
	}), &info, nil
}

func (rpc *RPC) GetSigningInfo(
	valcons string,
	ctx context.Context,
) (*types.SigningInfoResponse, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("signing-info") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching validator signing info",
		trace.WithAttributes(attribute.String("valcons", valcons)),
	)
	defer span.End()

	url := fmt.Sprintf("%s/cosmos/slashing/v1beta1/signing_infos/%s", rpc.Chain.LCDEndpoint, valcons)

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
	if !rpc.Chain.QueryEnabled("slashing-params") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching slashing params",
	)
	defer span.End()

	url := rpc.Chain.LCDEndpoint + "/cosmos/slashing/v1beta1/params"

	var response *types.SlashingParamsResponse
	info, err := rpc.Get(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	return response, &info, nil
}

func (rpc *RPC) GetConsumerSoftOutOutThreshold(
	ctx context.Context,
) (float64, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("params") {
		return 0, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching soft opt-out threshold params",
	)
	defer span.End()

	var response *types.ParamsResponse
	info, err := rpc.Get(
		rpc.Chain.LCDEndpoint+"/cosmos/params/v1beta1/params?subspace=ccvconsumer&key=SoftOptOutThreshold",
		&response,
		childQuerierCtx,
	)
	if err != nil {
		return 0, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return 0, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	valueStripped := strings.ReplaceAll(response.Param.Value, "\"", "")
	value, err := strconv.ParseFloat(valueStripped, 64)
	if err != nil {
		info.Success = false
		return 0, &info, err
	}

	return value, &info, nil
}

func (rpc *RPC) GetStakingParams(
	ctx context.Context,
) (*types.StakingParamsResponse, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("staking-params") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching staking params",
	)
	defer span.End()

	host := rpc.Chain.LCDEndpoint
	if rpc.Chain.IsConsumer() {
		host = rpc.Chain.ProviderChainLCD
	}

	url := host + "/cosmos/staking/v1beta1/params"

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

	height, err := utils.GetBlockHeightFromHeader(header)
	if err != nil {
		return info, err
	}

	rpc.Mutex.Lock()
	rpc.LastHeight[url] = height
	rpc.Mutex.Unlock()

	rpc.Logger.Trace().
		Str("url", url).
		Int64("height", height).
		Msg("Got response at height")

	return info, err
}
