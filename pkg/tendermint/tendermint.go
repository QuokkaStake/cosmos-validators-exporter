package tendermint

import (
	"encoding/json"
	"fmt"
	"main/pkg/config"
	"net/http"
	"time"

	types2 "main/pkg/types"
	"main/pkg/utils"

	"github.com/cosmos/cosmos-sdk/types"
	distributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/rs/zerolog"
)

type RPC struct {
	Chain   string
	URL     string
	Timeout int
	Logger  zerolog.Logger
}

func NewRPC(chain config.Chain, timeout int, logger zerolog.Logger) *RPC {
	return &RPC{
		Chain:   chain.Name,
		URL:     chain.LCDEndpoint,
		Timeout: timeout,
		Logger:  logger.With().Str("component", "rpc").Logger(),
	}
}

func (rpc *RPC) GetValidator(address string) (*types2.ValidatorResponse, types2.QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s",
		rpc.URL,
		address,
	)

	var response *types2.ValidatorResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return nil, info, err
	}

	return response, info, nil
}

func (rpc *RPC) GetDelegationsCount(address string) (*types2.PaginationResponse, types2.QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/delegations?pagination.count_total=true&pagination.limit=1",
		rpc.URL,
		address,
	)

	var response *types2.PaginationResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return nil, info, err
	}

	return response, info, nil
}

func (rpc *RPC) GetUnbondsCount(address string) (*types2.PaginationResponse, types2.QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/unbonding_delegations?pagination.count_total=true&pagination.limit=1",
		rpc.URL,
		address,
	)

	var response *types2.PaginationResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return nil, info, err
	}

	return response, info, nil
}

func (rpc *RPC) GetSingleDelegation(validator, wallet string) (types2.Balance, types2.QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/delegations/%s",
		rpc.URL,
		validator,
		wallet,
	)

	var response types2.SingleDelegationResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return types2.Balance{}, info, err
	}

	return types2.Balance{
		Amount: utils.StrToFloat64(response.DelegationResponse.Balance.Amount),
		Denom:  response.DelegationResponse.Balance.Denom,
	}, info, nil
}

func (rpc *RPC) GetAllValidators() (*types2.ValidatorsResponse, types2.QueryInfo, error) {
	url := fmt.Sprintf("%s/cosmos/staking/v1beta1/validators?pagination.count_total=true&pagination.limit=1000", rpc.URL)

	var response *types2.ValidatorsResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return nil, info, err
	}

	return response, info, nil
}

func (rpc *RPC) GetValidatorCommission(address string) ([]types2.Balance, types2.QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/distribution/v1beta1/validators/%s/commission",
		rpc.URL,
		address,
	)

	var response *distributionTypes.QueryValidatorCommissionResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return []types2.Balance{}, info, err
	}

	return utils.Map(response.Commission.Commission, func(balance types.DecCoin) types2.Balance {
		return types2.Balance{
			Amount: balance.Amount.MustFloat64(),
			Denom:  balance.Denom,
		}
	}), info, nil
}

func (rpc *RPC) GetDelegatorRewards(validator, wallet string) ([]types2.Balance, types2.QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/distribution/v1beta1/delegators/%s/rewards/%s",
		rpc.URL,
		wallet,
		validator,
	)

	var response *distributionTypes.QueryDelegationRewardsResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return []types2.Balance{}, info, err
	}

	return utils.Map(response.Rewards, func(balance types.DecCoin) types2.Balance {
		return types2.Balance{
			Amount: balance.Amount.MustFloat64(),
			Denom:  balance.Denom,
		}
	}), info, nil
}

func (rpc *RPC) GetWalletBalance(wallet string) ([]types2.Balance, types2.QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/bank/v1beta1/balances/%s",
		rpc.URL,
		wallet,
	)

	var response types2.BalancesResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return []types2.Balance{}, info, err
	}

	return utils.Map(response.Balances, func(balance types2.BalanceInResponse) types2.Balance {
		return types2.Balance{
			Amount: utils.StrToFloat64(balance.Amount),
			Denom:  balance.Denom,
		}
	}), info, nil
}

func (rpc *RPC) GetSigningInfo(valcons string) (*types2.SigningInfoResponse, *types2.QueryInfo, error) {
	url := fmt.Sprintf("%s/cosmos/slashing/v1beta1/signing_infos/%s", rpc.URL, valcons)

	var response *types2.SigningInfoResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return nil, &info, err
	}

	return response, &info, nil
}

func (rpc *RPC) GetSlashingParams() (*types2.SlashingParamsResponse, *types2.QueryInfo, error) {
	url := fmt.Sprintf("%s/cosmos/slashing/v1beta1/params", rpc.URL)

	var response *types2.SlashingParamsResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return nil, &info, err
	}

	return response, &info, nil
}

func (rpc *RPC) GetStakingParams() (*types2.StakingParamsResponse, *types2.QueryInfo, error) {
	url := fmt.Sprintf("%s/cosmos/staking/v1beta1/params", rpc.URL)

	var response *types2.StakingParamsResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return nil, &info, err
	}

	return response, &info, nil
}

func (rpc *RPC) Get(url string, target interface{}) (types2.QueryInfo, error) {
	client := &http.Client{
		Timeout: time.Duration(rpc.Timeout) * time.Second,
	}
	start := time.Now()

	info := types2.QueryInfo{
		Chain:   rpc.Chain,
		URL:     url,
		Success: false,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return info, err
	}

	req.Header.Set("User-Agent", "cosmos-validators-exporter")

	rpc.Logger.Trace().Str("url", url).Msg("Doing a query...")

	res, err := client.Do(req)
	if err != nil {
		info.Duration = time.Since(start)
		rpc.Logger.Warn().Str("url", url).Err(err).Msg("Query failed")
		return info, err
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		info.Duration = time.Since(start)
		rpc.Logger.Warn().
			Str("url", url).
			Err(err).
			Int("status", res.StatusCode).
			Msg("Query returned bad HTTP code")
		return info, fmt.Errorf("bad HTTP code: %d", res.StatusCode)
	}

	info.Duration = time.Since(start)

	rpc.Logger.Debug().Str("url", url).Dur("duration", time.Since(start)).Msg("Query is finished")

	err = json.NewDecoder(res.Body).Decode(target)
	info.Success = (err == nil)

	return info, err
}
