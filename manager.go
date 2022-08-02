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
					info                 *ValidatorResponse
					validatorQueryInfo   QueryInfo
					validatorQueryError  error
					delegators           *PaginationResponse
					delegatorsCountQuery QueryInfo
					delegatorsCountError error
					rank                 uint64
					totalStake           float64
					validatorsQueryInfo  QueryInfo
					validatorsQueryError error

					selfDelegationAmount     Balance
					selfDelegationQuery      *QueryInfo
					selfDelegationQueryError error

					commission           []Balance
					commissionQuery      QueryInfo
					commissionQueryError error

					selfDelegationRewards           []Balance
					selfDelegationRewardsQuery      *QueryInfo
					selfDelegationRewardsQueryError error

					walletBalance           []Balance
					walletBalanceQuery      *QueryInfo
					walletBalanceQueryError error

					signingInfo           *SigningInfoResponse
					signingInfoQuery      *QueryInfo
					signingInfoQueryError error

					validatorInfo ValidatorInfo
				)

				internalWg.Add(1)
				go func() {
					info, validatorQueryInfo, validatorQueryError = rpc.GetValidator(address)

					if validatorQueryError == nil {
						valConsAddress, err := info.Validator.ConsensusPubkey.GetValConsAddress(chain.BechConsensusPrefix)
						if err != nil {
							signingInfoQueryError = err
						} else {
							signingInfo, signingInfoQuery, signingInfoQueryError = rpc.GetSigningInfo(valConsAddress)
						}
					}

					internalWg.Done()
				}()

				internalWg.Add(1)
				go func() {
					delegators, delegatorsCountQuery, delegatorsCountError = rpc.GetDelegationsCount(address)
					internalWg.Done()
				}()

				internalWg.Add(1)
				go func() {
					rank, totalStake, validatorsQueryInfo, validatorsQueryError = m.GetValidatorRankAndTotalStake(chain, address, rpc)
					internalWg.Done()
				}()

				internalWg.Add(1)
				go func() {
					selfDelegationAmount, selfDelegationQuery, selfDelegationQueryError = m.GetSelfDelegationsBalance(chain, address, rpc)
					internalWg.Done()
				}()

				internalWg.Add(1)
				go func() {
					commission, commissionQuery, commissionQueryError = rpc.GetValidatorCommission(address)
					internalWg.Done()
				}()

				internalWg.Add(1)
				go func() {
					selfDelegationRewards, selfDelegationRewardsQuery, selfDelegationRewardsQueryError = m.GetSelfDelegationRewards(chain, address, rpc)
					internalWg.Done()
				}()

				internalWg.Add(1)
				go func() {
					walletBalance, walletBalanceQuery, walletBalanceQueryError = m.GetWalletBalance(chain, address, rpc)
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

					if hasPrice {
						validatorInfo.SelfDelegationUSD = m.CalculatePrices([]Balance{selfDelegationAmount}, price, chain)
					}
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

				if commissionQueryError != nil {
					m.Logger.Error().
						Err(commissionQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying validator commission")
				} else {
					validatorInfo.Commission = commission
					if hasPrice {
						validatorInfo.CommissionUSD = m.CalculatePrices(commission, price, chain)
					}
				}

				if selfDelegationRewardsQueryError != nil {
					m.Logger.Error().
						Err(selfDelegationRewardsQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying validator self-delegation rewards")
				} else {
					validatorInfo.SelfDelegationRewards = selfDelegationRewards
					if hasPrice {
						validatorInfo.SelfDelegationRewardsUSD = m.CalculatePrices(selfDelegationRewards, price, chain)
					}
				}

				if walletBalanceQueryError != nil {
					m.Logger.Error().
						Err(walletBalanceQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying validator wallet balance")
				} else {
					validatorInfo.WalletBalance = walletBalance
					if hasPrice {
						validatorInfo.WalletBalanceUSD = m.CalculatePrices(walletBalance, price, chain)
					}
				}

				if signingInfoQueryError != nil {
					m.Logger.Error().
						Err(signingInfoQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying validator signing info")
				} else if signingInfo != nil {
					validatorInfo.MissedBlocksCount = StrToInt64(signingInfo.ValSigningInfo.MissedBlocksCounter)
					validatorInfo.IsTombstoned = signingInfo.ValSigningInfo.Tombstoned
					validatorInfo.JailedUntil = signingInfo.ValSigningInfo.JailedUntil
					validatorInfo.StartHeight = StrToInt64(signingInfo.ValSigningInfo.StartHeight)
					validatorInfo.IndexOffset = StrToInt64(signingInfo.ValSigningInfo.IndexOffset)
				}

				rpcQueries := []QueryInfo{
					validatorQueryInfo,
					delegatorsCountQuery,
					validatorsQueryInfo,
					commissionQuery,
				}
				if selfDelegationQuery != nil {
					rpcQueries = append(rpcQueries, *selfDelegationQuery)
				}
				if selfDelegationRewardsQuery != nil {
					rpcQueries = append(rpcQueries, *selfDelegationRewardsQuery)
				}
				if walletBalanceQuery != nil {
					rpcQueries = append(rpcQueries, *walletBalanceQuery)
				}
				if signingInfoQuery != nil {
					rpcQueries = append(rpcQueries, *signingInfoQuery)
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

func (m *Manager) GetSelfDelegationsBalance(chain Chain, address string, rpc *RPC) (Balance, *QueryInfo, error) {
	if chain.BechWalletPrefix == "" {
		return Balance{}, nil, nil
	}

	wallet, err := ChangeBech32Prefix(address, chain.BechWalletPrefix)
	if err != nil {
		m.Logger.Error().
			Err(err).
			Str("chain", chain.Name).
			Str("address", address).
			Msg("Error converting validator address")
		return Balance{}, nil, err
	}

	balance, queryInfo, err := rpc.GetSingleDelegation(address, wallet)
	if err != nil {
		m.Logger.Error().
			Err(err).
			Str("chain", chain.Name).
			Str("address", address).
			Msg("Error querying for validator self-delegation")
		return Balance{}, &queryInfo, err
	}

	return balance, &queryInfo, err
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

func (m *Manager) GetSelfDelegationRewards(chain Chain, address string, rpc *RPC) ([]Balance, *QueryInfo, error) {
	if chain.BechWalletPrefix == "" {
		return []Balance{}, nil, nil
	}

	wallet, err := ChangeBech32Prefix(address, chain.BechWalletPrefix)
	if err != nil {
		m.Logger.Error().
			Err(err).
			Str("chain", chain.Name).
			Str("address", address).
			Msg("Error converting validator address")
		return []Balance{}, nil, err
	}

	balances, queryInfo, err := rpc.GetDelegatorRewards(address, wallet)
	if err != nil {
		m.Logger.Error().
			Err(err).
			Str("chain", chain.Name).
			Str("address", address).
			Msg("Error querying for validator self-delegation rewards")
		return []Balance{}, &queryInfo, err
	}

	return balances, &queryInfo, err
}

func (m *Manager) GetWalletBalance(chain Chain, address string, rpc *RPC) ([]Balance, *QueryInfo, error) {
	if chain.BechWalletPrefix == "" {
		return []Balance{}, nil, nil
	}

	wallet, err := ChangeBech32Prefix(address, chain.BechWalletPrefix)
	if err != nil {
		m.Logger.Error().
			Err(err).
			Str("chain", chain.Name).
			Str("address", address).
			Msg("Error converting validator address")
		return []Balance{}, nil, err
	}

	balances, queryInfo, err := rpc.GetWalletBalance(wallet)
	if err != nil {
		m.Logger.Error().
			Err(err).
			Str("chain", chain.Name).
			Str("address", address).
			Msg("Error querying for validator wallet balance")
		return []Balance{}, &queryInfo, err
	}

	return balances, &queryInfo, err
}

func (m *Manager) CalculatePrices(balances []Balance, rate float64, chain Chain) float64 {
	var usdPriceTotal float64 = 0
	for _, balance := range balances {
		if balance.Denom == chain.BaseDenom {
			usdPriceTotal += m.CalculatePrice(balance.Amount, rate, chain.DenomCoefficient)
		}
	}

	return usdPriceTotal
}

func (m *Manager) CalculatePrice(amount float64, rate float64, coefficient int64) float64 {
	return amount * rate / float64(coefficient)
}
