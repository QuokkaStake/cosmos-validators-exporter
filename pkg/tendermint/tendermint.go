package tendermint

import (
	"fmt"
	"main/pkg/config"
	"main/pkg/http"
	"main/pkg/types"
	"main/pkg/utils"

	"github.com/rs/zerolog"
)

type RPC struct {
	Chain   config.Chain
	Client  *http.Client
	Timeout int
	Logger  zerolog.Logger
}

func NewRPC(chain config.Chain, timeout int, logger zerolog.Logger) *RPC {
	return &RPC{
		Chain:   chain,
		Client:  http.NewClient(&logger, chain.Name),
		Timeout: timeout,
		Logger:  logger.With().Str("component", "rpc").Logger(),
	}
}

func (rpc *RPC) GetValidator(address string) (*types.ValidatorResponse, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("validator") {
		return nil, nil, nil
	}

	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s",
		rpc.Chain.LCDEndpoint,
		address,
	)

	var response *types.ValidatorResponse
	info, err := rpc.Client.Get(url, &response)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.ValidatorResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetDelegationsCount(address string) (*types.PaginationResponse, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("delegations") {
		return nil, nil, nil
	}

	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/delegations?pagination.count_total=true&pagination.limit=1",
		rpc.Chain.LCDEndpoint,
		address,
	)

	var response *types.PaginationResponse
	info, err := rpc.Client.Get(url, &response)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.PaginationResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetUnbondsCount(address string) (*types.PaginationResponse, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("unbonds") {
		return nil, nil, nil
	}

	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/unbonding_delegations?pagination.count_total=true&pagination.limit=1",
		rpc.Chain.LCDEndpoint,
		address,
	)

	var response *types.PaginationResponse
	info, err := rpc.Client.Get(url, &response)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.PaginationResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetSingleDelegation(validator, wallet string) (*types.Amount, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("self-delegation") {
		return nil, nil, nil
	}

	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/delegations/%s",
		rpc.Chain.LCDEndpoint,
		validator,
		wallet,
	)

	var response types.SingleDelegationResponse
	info, err := rpc.Client.Get(url, &response)
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

func (rpc *RPC) GetAllValidators() (*types.ValidatorsResponse, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("validators") {
		return nil, nil, nil
	}

	url := fmt.Sprintf("%s/cosmos/staking/v1beta1/validators?pagination.count_total=true&pagination.limit=1000", rpc.Chain.LCDEndpoint)

	var response *types.ValidatorsResponse
	info, err := rpc.Client.Get(url, &response)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.ValidatorsResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetValidatorCommission(address string) ([]types.Amount, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("commission") {
		return nil, nil, nil
	}

	url := fmt.Sprintf(
		"%s/cosmos/distribution/v1beta1/validators/%s/commission",
		rpc.Chain.LCDEndpoint,
		address,
	)

	var response *types.CommissionResponse
	info, err := rpc.Client.Get(url, &response)
	if err != nil {
		return []types.Amount{}, &info, err
	}

	return utils.Map(response.Commission.Commission, func(amount types.ResponseAmount) types.Amount {
		return amount.ToAmount()
	}), &info, nil
}

func (rpc *RPC) GetDelegatorRewards(validator, wallet string) ([]types.Amount, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("rewards") {
		return nil, nil, nil
	}

	url := fmt.Sprintf(
		"%s/cosmos/distribution/v1beta1/delegators/%s/rewards/%s",
		rpc.Chain.LCDEndpoint,
		wallet,
		validator,
	)

	var response *types.RewardsResponse
	info, err := rpc.Client.Get(url, &response)
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

func (rpc *RPC) GetWalletBalance(wallet string) ([]types.Amount, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("balance") {
		return nil, nil, nil
	}

	url := fmt.Sprintf(
		"%s/cosmos/bank/v1beta1/balances/%s",
		rpc.Chain.LCDEndpoint,
		wallet,
	)

	var response types.BalancesResponse
	info, err := rpc.Client.Get(url, &response)
	if err != nil {
		return []types.Amount{}, &info, err
	}

	return utils.Map(response.Balances, func(amount types.ResponseAmount) types.Amount {
		return amount.ToAmount()
	}), &info, nil
}

func (rpc *RPC) GetSigningInfo(valcons string) (*types.SigningInfoResponse, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("signing-info") {
		return nil, nil, nil
	}

	url := fmt.Sprintf("%s/cosmos/slashing/v1beta1/signing_infos/%s", rpc.Chain.LCDEndpoint, valcons)

	var response *types.SigningInfoResponse
	info, err := rpc.Client.Get(url, &response)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.SigningInfoResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}

func (rpc *RPC) GetSlashingParams() (*types.SlashingParamsResponse, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("slashing-params") {
		return nil, nil, nil
	}

	url := fmt.Sprintf("%s/cosmos/slashing/v1beta1/params", rpc.Chain.LCDEndpoint)

	var response *types.SlashingParamsResponse
	info, err := rpc.Client.Get(url, &response)
	if err != nil {
		return nil, &info, err
	}

	return response, &info, nil
}

func (rpc *RPC) GetStakingParams() (*types.StakingParamsResponse, *types.QueryInfo, error) {
	if !rpc.Chain.QueryEnabled("staking-params") {
		return nil, nil, nil
	}

	url := fmt.Sprintf("%s/cosmos/staking/v1beta1/params", rpc.Chain.LCDEndpoint)

	var response *types.StakingParamsResponse
	info, err := rpc.Client.Get(url, &response)
	if err != nil {
		return nil, &info, err
	}

	if response.Code != 0 {
		info.Success = false
		return &types.StakingParamsResponse{}, &info, fmt.Errorf("expected code 0, but got %d", response.Code)
	}

	return response, &info, nil
}
