package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	distributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/rs/zerolog"
)

type RPC struct {
	URL     string
	Timeout int
	Logger  zerolog.Logger
}

func NewRPC(url string, timeout int, logger zerolog.Logger) *RPC {
	return &RPC{
		URL:     url,
		Timeout: timeout,
		Logger:  logger.With().Str("component", "rpc").Logger(),
	}
}

func (rpc *RPC) GetValidator(address string) (*ValidatorResponse, QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s",
		rpc.URL,
		address,
	)

	var response *ValidatorResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return nil, info, err
	}

	return response, info, nil
}

func (rpc *RPC) GetDelegationsCount(address string) (*PaginationResponse, QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/delegations?pagination.count_total=true&pagination.limit=1",
		rpc.URL,
		address,
	)

	var response *PaginationResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return nil, info, err
	}

	return response, info, nil
}

func (rpc *RPC) GetUnbondsCount(address string) (*PaginationResponse, QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/unbonding_delegations?pagination.count_total=true&pagination.limit=1",
		rpc.URL,
		address,
	)

	var response *PaginationResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return nil, info, err
	}

	return response, info, nil
}

func (rpc *RPC) GetSingleDelegation(validator, wallet string) (Balance, QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/delegations/%s",
		rpc.URL,
		validator,
		wallet,
	)

	var response SingleDelegationResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return Balance{}, info, err
	}

	return Balance{
		Amount: StrToFloat64(response.DelegationResponse.Balance.Amount),
		Denom:  response.DelegationResponse.Balance.Denom,
	}, info, nil
}

func (rpc *RPC) GetAllValidators() (*ValidatorsResponse, QueryInfo, error) {
	url := fmt.Sprintf("%s/cosmos/staking/v1beta1/validators?pagination.count_total=true&pagination.limit=1000", rpc.URL)

	var response *ValidatorsResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return nil, info, err
	}

	return response, info, nil
}

func (rpc *RPC) GetValidatorCommission(address string) ([]Balance, QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/distribution/v1beta1/validators/%s/commission",
		rpc.URL,
		address,
	)

	var response *distributionTypes.QueryValidatorCommissionResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return []Balance{}, info, err
	}

	return Map(response.Commission.Commission, func(balance types.DecCoin) Balance {
		return Balance{
			Amount: balance.Amount.MustFloat64(),
			Denom:  balance.Denom,
		}
	}), info, nil
}

func (rpc *RPC) GetDelegatorRewards(validator, wallet string) ([]Balance, QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/distribution/v1beta1/delegators/%s/rewards/%s",
		rpc.URL,
		wallet,
		validator,
	)

	var response *distributionTypes.QueryDelegationRewardsResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return []Balance{}, info, err
	}

	return Map(response.Rewards, func(balance types.DecCoin) Balance {
		return Balance{
			Amount: balance.Amount.MustFloat64(),
			Denom:  balance.Denom,
		}
	}), info, nil
}

func (rpc *RPC) GetWalletBalance(wallet string) ([]Balance, QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/bank/v1beta1/balances/%s",
		rpc.URL,
		wallet,
	)

	var response BalancesResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return []Balance{}, info, err
	}

	return Map(response.Balances, func(balance BalanceInResponse) Balance {
		return Balance{
			Amount: StrToFloat64(balance.Amount),
			Denom:  balance.Denom,
		}
	}), info, nil
}

func (rpc *RPC) GetSigningInfo(valcons string) (*SigningInfoResponse, *QueryInfo, error) {
	url := fmt.Sprintf("%s/cosmos/slashing/v1beta1/signing_infos/%s", rpc.URL, valcons)

	var response *SigningInfoResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return nil, &info, err
	}

	return response, &info, nil
}

func (rpc *RPC) GetSlashingParams() (*SlashingParamsResponse, *QueryInfo, error) {
	url := fmt.Sprintf("%s/cosmos/slashing/v1beta1/params", rpc.URL)

	var response *SlashingParamsResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return nil, &info, err
	}

	return response, &info, nil
}

func (rpc *RPC) Get(url string, target interface{}) (QueryInfo, error) {
	client := &http.Client{
		Timeout: time.Duration(rpc.Timeout) * time.Second,
	}
	start := time.Now()

	info := QueryInfo{
		URL:     url,
		Success: false,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return info, err
	}

	rpc.Logger.Trace().Str("url", url).Msg("Doing a query...")

	res, err := client.Do(req)
	if err != nil {
		info.Duration = time.Since(start)
		rpc.Logger.Warn().Str("url", url).Err(err).Msg("Query failed")
		return info, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
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
