package main

import (
	"time"
)

type ValidatorResponse struct {
	Validator Validator `json:"validator"`
}

type Validator struct {
	OperatorAddress   string               `json:"validator_address"`
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
	TokensUSD               float64
	Jailed                  bool
	Status                  string
	CommissionRate          float64
	CommissionMaxRate       float64
	CommissionMaxChangeRate float64
	CommissionUpdateTime    time.Time
	UnbondingHeight         int64
	UnbondingTime           time.Time
	MinSelfDelegation       int64
	DelegatorsCount         int64
	SelfDelegation          float64
	SelfDelegationUSD       float64
}

func NewValidatorInfo(validator Validator) ValidatorInfo {
	return ValidatorInfo{
		Address:                 validator.OperatorAddress,
		Moniker:                 validator.Description.Moniker,
		Identity:                validator.Description.Identity,
		Website:                 validator.Description.Website,
		SecurityContact:         validator.Description.SecurityContact,
		Details:                 validator.Description.Details,
		Tokens:                  StrToFloat64(validator.Tokens),
		Jailed:                  validator.Jailed,
		Status:                  validator.Status,
		CommissionRate:          StrToFloat64(validator.Commission.CommissionRates.Rate),
		CommissionMaxRate:       StrToFloat64(validator.Commission.CommissionRates.MaxRate),
		CommissionMaxChangeRate: StrToFloat64(validator.Commission.CommissionRates.MaxChangeRate),
		CommissionUpdateTime:    validator.Commission.UpdateTime,
		UnbondingHeight:         StrToInt64(validator.UnbondingHeight),
		UnbondingTime:           validator.UnbondingTime,
		MinSelfDelegation:       StrToInt64(validator.MinSelfDelegation),
	}
}

type ValidatorQuery struct {
	Chain   string
	Address string
	Queries []QueryInfo
	Info    *ValidatorInfo
}

type PaginationResponse struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Total string `json:"total"`
}

type QueryInfo struct {
	URL      string
	Duration time.Duration
	Success  bool
}

func (q *ValidatorQuery) GetSuccessfulQueriesCount() int64 {
	var count int64 = 0

	for _, query := range q.Queries {
		if query.Success {
			count++
		}
	}

	return count
}
