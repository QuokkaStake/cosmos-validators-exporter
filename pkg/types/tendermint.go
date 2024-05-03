package types

import (
	"main/pkg/constants"
	"main/pkg/utils"
	"time"
)

type ValidatorResponse struct {
	Code      int       `json:"code"`
	Validator Validator `json:"validator"`
}

type Validator struct {
	OperatorAddress string          `json:"operator_address"`
	ConsensusPubkey ConsensusPubkey `json:"consensus_pubkey"`
	Jailed          bool            `json:"jailed"`
	Status          string          `json:"status"`
	Tokens          string          `json:"tokens"`
	DelegatorShares string          `json:"delegator_shares"`
	Description     struct {
		Moniker         string `json:"moniker"`
		Identity        string `json:"identity"`
		Website         string `json:"website"`
		SecurityContact string `json:"security_contact"`
		Details         string `json:"details"`
	} `json:"description"`
	UnbondingHeight string    `json:"unbonding_height"`
	UnbondingTime   time.Time `json:"unbonding_time"`
	Commission      struct {
		CommissionRates struct {
			Rate          string `json:"rate"`
			MaxRate       string `json:"max_rate"`
			MaxChangeRate string `json:"max_change_rate"`
		} `json:"commission_rates"`
		UpdateTime time.Time `json:"update_time"`
	} `json:"commission"`
	MinSelfDelegation string `json:"min_self_delegation"`
}

func (v Validator) Active() bool {
	return v.Status == constants.ValidatorStatusBonded
}

type ConsensusPubkey struct {
	Type string `json:"@type"`
	Key  string `json:"key"`
}

type PaginationResponse struct {
	Code       int        `json:"code"`
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Total string `json:"total"`
}

type ValidatorsResponse struct {
	Code       int         `json:"code"`
	Validators []Validator `json:"validators"`
}

type BalancesResponse struct {
	Code     int              `json:"code"`
	Balances []ResponseAmount `json:"balances"`
}

type ResponseAmount struct {
	Amount string `json:"amount"`
	Denom  string `json:"denom"`
}

func (a ResponseAmount) ToAmount() Amount {
	return Amount{
		Amount: utils.StrToFloat64(a.Amount),
		Denom:  a.Denom,
	}
}

type SigningInfoResponse struct {
	Code           int `json:"code"`
	ValSigningInfo struct {
		Address             string    `json:"address"`
		StartHeight         string    `json:"start_height"`
		IndexOffset         string    `json:"index_offset"`
		JailedUntil         time.Time `json:"jailed_until"`
		Tombstoned          bool      `json:"tombstoned"`
		MissedBlocksCounter string    `json:"missed_blocks_counter"`
	} `json:"val_signing_info"`
}

type SingleDelegationResponse struct {
	Code               int                `json:"code"`
	DelegationResponse DelegationResponse `json:"delegation_response"`
}

type DelegationResponse struct {
	Balance ResponseAmount `json:"balance"`
}

type RewardsResponse struct {
	Code    int              `json:"code"`
	Rewards []ResponseAmount `json:"rewards"`
}

type CommissionResponse struct {
	Code       int `json:"code"`
	Commission struct {
		Commission []ResponseAmount `json:"commission"`
	} `json:"commission"`
}

type ParamsResponse struct {
	Code  int `json:"code"`
	Param struct {
		Subspace string `json:"subspace"`
		Key      string `json:"key"`
		Value    string `json:"value"`
	} `json:"param"`
}
