package tendermint

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"main/pkg/config"
	"main/pkg/http"
	"main/pkg/types"
	"main/pkg/utils"
	"strconv"
	"strings"
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/proto"

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

	Registry   codecTypes.InterfaceRegistry
	ParseCodec *codec.ProtoCodec

	LastHeight map[string]int64
	Mutex      sync.Mutex
}

func NewRPC(chain config.Chain, timeout int, logger zerolog.Logger, tracer trace.Tracer) *RPC {
	interfaceRegistry := codecTypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	parseCodec := codec.NewProtoCodec(interfaceRegistry)

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
		Registry:   interfaceRegistry,
		ParseCodec: parseCodec,
	}
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

	if response.Code != 0 {
		info.Success = false
		return nil, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return utils.Map(response.Balances, func(amount types.ResponseAmount) types.Amount {
		return amount.ToAmount()
	}), &info, nil
}

func (rpc *RPC) GetSigningInfo(
	valcons string,
	ctx context.Context,
) (*slashingTypes.ValidatorSigningInfo, *types.QueryInfo, error) {
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

	var response slashingTypes.QuerySigningInfoResponse
	info, err := rpc.Get2(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	return &response.ValSigningInfo, &info, nil
}

func (rpc *RPC) GetSlashingParams(
	ctx context.Context,
) (*slashingTypes.Params, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("slashing-params") {
		return nil, nil, nil
	}

	childQuerierCtx, span := rpc.Tracer.Start(
		ctx,
		"Fetching slashing params",
	)
	defer span.End()

	url := rpc.Chain.LCDEndpoint + "/cosmos/slashing/v1beta1/params"

	var response slashingTypes.QueryParamsResponse
	info, err := rpc.Get2(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	return &response.Params, &info, nil
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
) (*stakingTypes.Params, *types.QueryInfo, error) {
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

	var response stakingTypes.QueryParamsResponse
	info, err := rpc.Get2(url, &response, childQuerierCtx)
	if err != nil {
		return nil, &info, err
	}

	return &response.Params, &info, nil
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

	body, header, info, err := rpc.Client.Get(
		url,
		types.HTTPPredicateCheckHeightAfter(previousHeight),
		ctx,
	)

	if err != nil {
		return info, err
	}

	if unmarshalErr := json.Unmarshal(body, target); unmarshalErr != nil {
		rpc.Logger.Warn().Str("url", url).Err(unmarshalErr).Msg("JSON unmarshalling failed")
		return info, unmarshalErr
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

func (rpc *RPC) Get2(
	url string,
	target proto.Message,
	ctx context.Context,
) (types.QueryInfo, error) {
	rpc.Mutex.Lock()
	previousHeight, found := rpc.LastHeight[url]
	if !found {
		previousHeight = 0
	}
	rpc.Mutex.Unlock()

	body, header, info, err := rpc.Client.Get(
		url,
		types.HTTPPredicateCheckHeightAfter(previousHeight),
		ctx,
	)

	if err != nil {
		return info, err
	}

	// check whether the response is error first
	var errorResponse types.LCDError
	if err := json.Unmarshal(body, &errorResponse); err == nil {
		// if we successfully unmarshalled it into LCDError, so err == nil,
		// that means the response is indeed an error.
		if errorResponse.Code != 0 {
			rpc.Logger.Warn().Str("url", url).
				Err(err).
				Int("code", errorResponse.Code).
				Str("message", errorResponse.Message).
				Msg("LCD request returned an error")
			info.Success = false
			return info, errors.New(errorResponse.Message)
		}
	}

	if unmarshalErr := rpc.ParseCodec.UnmarshalJSON(body, target); unmarshalErr != nil {
		rpc.Logger.Warn().Str("url", url).Err(unmarshalErr).Msg("JSON unmarshalling failed")
		return info, unmarshalErr
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
