package main

import (
	"sort"
	"sync"

	"github.com/rs/zerolog"
)

type Manager struct {
	Config    Config
	Coingecko *Coingecko
	Logger    zerolog.Logger
}

func NewManager(config Config, logger *zerolog.Logger) *Manager {
	return &Manager{
		Config:    config,
		Coingecko: NewCoingecko(logger),
		Logger:    logger.With().Str("component", "manager").Logger(),
	}
}

func (m *Manager) GetAllValidators() []ValidatorQuery {
	currenciesList := m.Config.GetCoingeckoCurrencies()
	currenciesRates := m.Coingecko.FetchPrices(currenciesList)

	length := 0
	for _, chain := range m.Config.Chains {
		for range chain.Validators {
			length++
		}
	}

	validators := make([]ValidatorQuery, length)

	var wg sync.WaitGroup
	wg.Add(length)

	index := 0

	for _, chain := range m.Config.Chains {
		rpc := NewRPC(chain.LCDEndpoint, m.Config.Timeout, m.Logger)

		for _, address := range chain.Validators {
			go func(address string, chain Chain, index int) {
				defer wg.Done()

				var internalWg sync.WaitGroup

				var (
					info                     *ValidatorResponse
					validatorQueryInfo       QueryInfo
					validatorQueryError      error
					delegators               *PaginationResponse
					delegatorsCountQuery     QueryInfo
					delegatorsCountError     error
					rank                     uint64
					totalStake               float64
					validatorsQueryInfo      QueryInfo
					validatorsQueryError     error
					selfDelegationAmount     float64
					selfDelegationQuery      *QueryInfo
					selfDelegationQueryError error

					validatorInfo ValidatorInfo
				)

				internalWg.Add(4)
				go func() {
					info, validatorQueryInfo, validatorQueryError = rpc.GetValidator(address)
					internalWg.Done()
				}()

				go func() {
					delegators, delegatorsCountQuery, delegatorsCountError = rpc.GetDelegationsCount(address)
					internalWg.Done()
				}()

				go func() {
					rank, totalStake, validatorsQueryInfo, validatorsQueryError = m.GetValidatorRankAndTotalStake(chain, address, rpc)
					internalWg.Done()
				}()

				go func() {
					selfDelegationAmount, selfDelegationQuery, selfDelegationQueryError = m.GetSelfDelegationsBalance(chain, address, rpc)
					internalWg.Done()
				}()

				internalWg.Wait()

				if validatorQueryError != nil {
					m.Logger.Error().
						Err(validatorQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying validator")
					validatorInfo = ValidatorInfo{}
				} else {
					validatorInfo = NewValidatorInfo(info.Validator)
				}

				price, hasPrice := currenciesRates[chain.CoingeckoCurrency]
				if hasPrice {
					validatorInfo.TokensUSD = m.CalculatePrice(validatorInfo.Tokens, price, chain.DenomCoefficient)
				}

				if delegatorsCountError != nil {
					m.Logger.Error().
						Err(delegatorsCountError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying validator delegations count")
				} else {
					validatorInfo.DelegatorsCount = StrToInt64(delegators.Pagination.Total)
				}

				if selfDelegationQueryError != nil {
					m.Logger.Error().
						Err(selfDelegationQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying self-delegations for validator")
				} else {
					validatorInfo.SelfDelegation = selfDelegationAmount
					validatorInfo.SelfDelegationUSD = m.CalculatePrice(
						selfDelegationAmount,
						price,
						chain.DenomCoefficient,
					)
				}

				if validatorsQueryError != nil {
					m.Logger.Error().
						Err(validatorsQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying validators list")
				} else {
					validatorInfo.Rank = rank
					validatorInfo.TotalStake = totalStake
				}

				rpcQueries := []QueryInfo{
					validatorQueryInfo,
					delegatorsCountQuery,
					validatorsQueryInfo,
				}
				if selfDelegationQuery != nil {
					rpcQueries = append(rpcQueries, *selfDelegationQuery)
				}

				query := ValidatorQuery{
					Chain:   chain.Name,
					Address: address,
					Queries: rpcQueries,
					Info:    validatorInfo,
				}

				validators[index] = query
			}(address, chain, index)

			index++
		}
	}

	wg.Wait()

	return validators
}

func (m *Manager) GetSelfDelegationsBalance(chain Chain, address string, rpc *RPC) (float64, *QueryInfo, error) {
	if chain.BechWalletPrefix == "" {
		return 0, nil, nil
	}

	wallet, err := ChangeBech32Prefix(address, chain.BechWalletPrefix)
	if err != nil {
		m.Logger.Error().
			Err(err).
			Str("chain", chain.Name).
			Str("address", address).
			Msg("Error converting validator address")
		return 0, nil, err
	}

	balance, queryInfo, err := rpc.GetSingleDelegation(address, wallet)
	if err != nil {
		m.Logger.Error().
			Err(err).
			Str("chain", chain.Name).
			Str("address", address).
			Msg("Error querying for validator self-delegation")
		return 0, &queryInfo, err
	}

	if balance.DelegationResponse == nil {
		return 0, &queryInfo, err
	}

	return balance.DelegationResponse.Delegation.Shares.MustFloat64(), &queryInfo, err
}

func (m *Manager) GetValidatorRankAndTotalStake(chain Chain, address string, rpc *RPC) (uint64, float64, QueryInfo, error) {
	allValidators, info, err := rpc.GetAllValidators()
	if err != nil {
		m.Logger.Error().
			Err(err).
			Str("chain", chain.Name).
			Str("address", address).
			Msg("Error querying for validatos")
		return 0, 0, info, err
	}

	activeValidators := Filter(allValidators.Validators, func(v Validator) bool {
		return v.Status == "BOND_STATUS_BONDED"
	})

	sort.Slice(activeValidators, func(i, j int) bool {
		return StrToFloat64(activeValidators[i].DelegatorShares) > StrToFloat64(activeValidators[j].DelegatorShares)
	})

	var validatorRank uint64 = 0
	var totalStake float64 = 0

	for index, validator := range activeValidators {
		totalStake += StrToFloat64(validator.DelegatorShares)
		if validator.OperatorAddress == address {
			validatorRank = uint64(index) + 1
		}
	}

	return validatorRank, totalStake, info, nil
}

// func (m *Manager) MaybeGetUsdPrice(
// 	chain Chain,
// 	balances []types.Coin,
// 	rates map[string]float64,
// ) float64 {
// 	price, hasPrice := rates[chain.CoingeckoCurrency]
// 	if !hasPrice {
// 		return 0
// 	}

// 	var usdPriceTotal float64 = 0
// 	for _, balance := range balances {
// 		if balance.Denom == chain.BaseDenom {
// 			usdPriceTotal += m.CalculatePrice(balance)
// 			usdPriceTotal += StrToFloat64(balance.Amount) * price / float64(chain.DenomCoefficient)
// 		}
// 	}

// 	return usdPriceTotal
// }

func (m *Manager) CalculatePrice(amount float64, rate float64, coefficient int64) float64 {
	return amount * rate / float64(coefficient)
}
