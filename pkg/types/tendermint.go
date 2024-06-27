package types

import (
	"main/pkg/constants"
	"time"

	"cosmossdk.io/math"
)

type ValidatorResponse struct {
	Code      int       `json:"code"`
	Validator Validator `json:"validator"`
}

type ValidatorDescription struct {
	Moniker         string `json:"moniker"`
	Identity        string `json:"identity"`
	Website         string `json:"website"`
	SecurityContact string `json:"security_contact"`
	Details         string `json:"details"`
}

type ValidatorCommission struct {
	CommissionRates ValidatorCommissionRates `json:"commission_rates"`
	UpdateTime      time.Time                `json:"update_time"`
}

type ValidatorCommissionRates struct {
	Rate          math.LegacyDec `json:"rate"`
	MaxRate       math.LegacyDec `json:"max_rate"`
	MaxChangeRate math.LegacyDec `json:"max_change_rate"`
}

type Validator struct {
	OperatorAddress   string               `json:"operator_address"`
	ConsensusPubkey   ConsensusPubkey      `json:"consensus_pubkey"`
	Jailed            bool                 `json:"jailed"`
	Status            string               `json:"status"`
	Tokens            string               `json:"tokens"`
	DelegatorShares   math.LegacyDec       `json:"delegator_shares"`
	Description       ValidatorDescription `json:"description"`
	UnbondingHeight   string               `json:"unbonding_height"`
	UnbondingTime     time.Time            `json:"unbonding_time"`
	Commission        ValidatorCommission  `json:"commission"`
	MinSelfDelegation string               `json:"min_self_delegation"`
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
	Total uint64 `json:"total,string"`
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
	Amount math.LegacyDec `json:"amount"`
	Denom  string         `json:"denom"`
}

func (a ResponseAmount) ToAmount() Amount {
	return Amount{
		Amount: a.Amount.MustFloat64(),
		Denom:  a.Denom,
	}
}

type SigningInfo struct {
	Address             string    `json:"address"`
	StartHeight         string    `json:"start_height"`
	IndexOffset         string    `json:"index_offset"`
	JailedUntil         time.Time `json:"jailed_until"`
	Tombstoned          bool      `json:"tombstoned"`
	MissedBlocksCounter math.Int  `json:"missed_blocks_counter"`
}

type SigningInfoResponse struct {
	Code           int         `json:"code"`
	ValSigningInfo SigningInfo `json:"val_signing_info"`
}

type AssignedKeyResponse struct {
	Code            int    `json:"code"`
	ConsumerAddress string `json:"consumer_address"`
}

type SlashingParams struct {
	SignedBlocksWindow math.Int `json:"signed_blocks_window"`
}

type SlashingParamsResponse struct {
	Code           int            `json:"code"`
	SlashingParams SlashingParams `json:"params"`
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

type StakingParams struct {
	MaxValidators int `json:"max_validators"`
}

type StakingParamsResponse struct {
	Code          int           `json:"code"`
	StakingParams StakingParams `json:"params"`
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

type DefaultNodeInfo struct {
	Network string `json:"network"`
	Version string `json:"version"`
}

type ApplicationVersion struct {
	Name             string `json:"name"`
	AppName          string `json:"app_name"`
	Version          string `json:"version"`
	CosmosSDKVersion string `json:"cosmos_sdk_version"`
}

type NodeInfoResponse struct {
	Code               int                `json:"code"`
	DefaultNodeInfo    DefaultNodeInfo    `json:"default_node_info"`
	ApplicationVersion ApplicationVersion `json:"application_version"`
}

type ConsumerValidator struct {
	ProviderAddress string `json:"provider_address"`
}

type ConsumerValidatorsResponse struct {
	Code       int                 `json:"code"`
	Validators []ConsumerValidator `json:"validators"`
}

type ConsumerChainInfo struct {
	ChainID        string   `json:"chain_id"`
	TopN           int      `json:"top_n"`
	MinPowerInTopN math.Int `json:"min_power_in_top_N"`
}

type ConsumerInfoResponse struct {
	Code   int                 `json:"code"`
	Chains []ConsumerChainInfo `json:"chains"`
}

type ValidatorConsumerChains struct {
	Code             int      `json:"code"`
	ConsumerChainIds []string `json:"consumer_chain_ids"`
}

type ConsumerCommissionResponse struct {
	Code int            `json:"code"`
	Rate math.LegacyDec `json:"rate"`
}
