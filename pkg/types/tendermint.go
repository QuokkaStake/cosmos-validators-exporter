package types

import (
	b64 "encoding/base64"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptoTypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types"
	"main/pkg/utils"
	"time"
)

type ValidatorResponse struct {
	Code      int       `json:"code"`
	Validator Validator `json:"validator"`
}

type Validator struct {
	OperatorAddress   string               `json:"operator_address"`
	ConsensusPubkey   ConsensusPubkey      `json:"consensus_pubkey"`
	Jailed            bool                 `json:"jailed"`
	Status            string               `json:"status"`
	Tokens            string               `json:"tokens"`
	DelegatorShares   string               `json:"delegator_shares"`
	Description       ValidatorDescription `json:"description"`
	UnbondingHeight   string               `json:"unbonding_height"`
	UnbondingTime     time.Time            `json:"unbonding_time"`
	Commission        ValidatorCommission  `json:"commission"`
	MinSelfDelegation string               `json:"min_self_delegation"`
}

type ConsensusPubkey struct {
	Type string `json:"@type"`
	Key  string `json:"key"`
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
	Rate          string `json:"rate"`
	MaxRate       string `json:"max_rate"`
	MaxChangeRate string `json:"max_change_rate"`
}

type ValidatorInfo struct {
	Address                 string
	Moniker                 string
	Identity                string
	Website                 string
	SecurityContact         string
	Details                 string
	Tokens                  float64
	Jailed                  bool
	Status                  string
	CommissionRate          float64
	CommissionMaxRate       float64
	CommissionMaxChangeRate float64
}

func (key *ConsensusPubkey) GetValConsAddress(prefix string) (string, error) {
	encCfg := simapp.MakeTestEncodingConfig()
	interfaceRegistry := encCfg.InterfaceRegistry

	sDec, _ := b64.StdEncoding.DecodeString(key.Key)
	pk := codecTypes.Any{
		TypeUrl: key.Type,
		Value:   append([]byte{10, 32}, sDec...),
	}

	var pkProto cryptoTypes.PubKey
	if err := interfaceRegistry.UnpackAny(&pk, &pkProto); err != nil {
		return "", err
	}

	cosmosValCons := types.ConsAddress(pkProto.Address()).String()
	properValCons, err := utils.ChangeBech32Prefix(cosmosValCons, prefix)
	if err != nil {
		return "", err
	}

	return properValCons, nil
}

type ValidatorQuery struct {
	Chain   string
	Address string
	Queries []QueryInfo
	Info    ValidatorInfo
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
	Code           int                  `json:"code"`
	ValSigningInfo ValidatorSigningInfo `json:"val_signing_info"`
}

type ValidatorSigningInfo struct {
	Address             string    `json:"address"`
	StartHeight         string    `json:"start_height"`
	IndexOffset         string    `json:"index_offset"`
	JailedUntil         time.Time `json:"jailed_until"`
	Tombstoned          bool      `json:"tombstoned"`
	MissedBlocksCounter string    `json:"missed_blocks_counter"`
}

type SlashingParamsResponse struct {
	SlashingParams SlashingParams `json:"params"`
}

type SlashingParams struct {
	SignedBlocksWindow string `json:"signed_blocks_window"`
}

type SingleDelegationResponse struct {
	Code               int                `json:"code"`
	DelegationResponse DelegationResponse `json:"delegation_response"`
}

type DelegationResponse struct {
	Balance ResponseAmount `json:"balance"`
}

type StakingParams struct {
	MaxValidators int `json:"max_validators"`
}

type StakingParamsResponse struct {
	StakingParams StakingParams `json:"params"`
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
