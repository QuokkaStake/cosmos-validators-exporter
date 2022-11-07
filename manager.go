package main

import (
	"sort"
	"sync"

	"github.com/rs/zerolog"
)

type Manager struct {
	Config      Config
	Coingecko   *Coingecko
	DexScreener *DexScreener
	Logger      zerolog.Logger
}

func NewManager(config Config, logger *zerolog.Logger) *Manager {
	return &Manager{
		Config:      config,
		Coingecko:   NewCoingecko(logger),
		DexScreener: NewDexScreener(logger),
		Logger:      logger.With().Str("component", "manager").Logger(),
	}
}

func (m *Manager) GetCurrencies() map[string]float64 {
	currenciesList := m.Config.GetCoingeckoCurrencies()
	currenciesRates := m.Coingecko.FetchPrices(currenciesList)

	currenciesRatesToChains := map[string]float64{}
	for _, chain := range m.Config.Chains {
		// using coingeckon response
		if rate, ok := currenciesRates[chain.CoingeckoCurrency]; ok {
			currenciesRatesToChains[chain.Name] = rate
			continue
		}

		// using dexscreener response
		if chain.DexScreenerChainID != "" && chain.DexScreenerPair != "" {
			rate, err := m.DexScreener.GetCurrency(chain.DexScreenerChainID, chain.DexScreenerPair)
			if err == nil {
				currenciesRatesToChains[chain.Name] = rate
			}
		}
	}

	return currenciesRatesToChains
}

func (m *Manager) GetAllValidators() []ValidatorQuery {
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
					rank                 uint64
					totalValidators      int
					totalStake           float64
					lastValidatorStake   float64
					validatorsQueryInfo  QueryInfo
					validatorsQueryError error

					delegators           *PaginationResponse
					delegatorsCountQuery QueryInfo
					delegatorsCountError error

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

					slashingParams           *SlashingParamsResponse
					slashingParamsQuery      *QueryInfo
					slashingParamsQueryError error

					stakingParams           *StakingParamsResponse
					stakingParamsQuery      *QueryInfo
					stakingParamsQueryError error

					unbonds                *PaginationResponse
					unbondsCountQuery      QueryInfo
					unbondsCountQueryError error

					validatorInfo ValidatorInfo
				)

				internalWg.Add(1)
				go func() {
					info, validatorQueryInfo, validatorQueryError = rpc.GetValidator(address)

					if validatorQueryError == nil && chain.BechConsensusPrefix != "" {
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
					rank, totalValidators, totalStake, lastValidatorStake, validatorsQueryInfo, validatorsQueryError = m.GetValidatorRankAndTotalStake(chain, address, rpc)
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

				internalWg.Add(1)
				go func() {
					slashingParams, slashingParamsQuery, slashingParamsQueryError = rpc.GetSlashingParams()
					internalWg.Done()
				}()

				internalWg.Add(1)
				go func() {
					stakingParams, stakingParamsQuery, stakingParamsQueryError = rpc.GetStakingParams()
					internalWg.Done()
				}()

				internalWg.Add(1)
				go func() {
					unbonds, unbondsCountQuery, unbondsCountQueryError = rpc.GetUnbondsCount(address)
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
				}

				if validatorsQueryError != nil {
					m.Logger.Error().
						Err(validatorsQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying validators list")
				} else {
					validatorInfo.Rank = rank
					validatorInfo.TotalValidators = totalValidators
					validatorInfo.TotalStake = totalStake

					// should be 0 if there are not enough validators
					if slashingParams != nil && totalValidators >= stakingParams.StakingParams.MaxValidators {
						validatorInfo.LastValidatorStake = lastValidatorStake
					}
				}

				if commissionQueryError != nil {
					m.Logger.Error().
						Err(commissionQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying validator commission")
				} else {
					validatorInfo.Commission = commission
				}

				if selfDelegationRewardsQueryError != nil {
					m.Logger.Error().
						Err(selfDelegationRewardsQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying validator self-delegation rewards")
				} else {
					validatorInfo.SelfDelegationRewards = selfDelegationRewards
				}

				if walletBalanceQueryError != nil {
					m.Logger.Error().
						Err(walletBalanceQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying validator wallet balance")
				} else {
					validatorInfo.WalletBalance = walletBalance
				}

				if signingInfoQueryError != nil {
					m.Logger.Error().
						Err(signingInfoQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying validator signing info")
				} else if signingInfo != nil && signingInfo.ValSigningInfo.Address != "" {
					validatorInfo.MissedBlocksCount = StrToInt64(signingInfo.ValSigningInfo.MissedBlocksCounter)
					validatorInfo.IsTombstoned = signingInfo.ValSigningInfo.Tombstoned
					validatorInfo.JailedUntil = signingInfo.ValSigningInfo.JailedUntil
					validatorInfo.StartHeight = StrToInt64(signingInfo.ValSigningInfo.StartHeight)
					validatorInfo.IndexOffset = StrToInt64(signingInfo.ValSigningInfo.IndexOffset)
				}

				if slashingParamsQueryError != nil {
					m.Logger.Error().
						Err(slashingParamsQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying slashing params")
				} else if slashingParams != nil && slashingParams.SlashingParams.SignedBlocksWindow != "" {
					validatorInfo.SignedBlocksWindow = StrToInt64(slashingParams.SlashingParams.SignedBlocksWindow)
				}

				if stakingParamsQueryError != nil {
					m.Logger.Error().
						Err(stakingParamsQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying staking params")
				} else if stakingParams != nil {
					validatorInfo.ActiveValidatorsCount = int64(stakingParams.StakingParams.MaxValidators)
				}

				if unbondsCountQueryError != nil {
					m.Logger.Error().
						Err(unbondsCountQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying unbonding delegations count")
				} else if unbonds != nil {
					validatorInfo.UnbondsCount = StrToInt64(unbonds.Pagination.Total)
				}

				rpcQueries := []QueryInfo{
					validatorQueryInfo,
					delegatorsCountQuery,
					unbondsCountQuery,
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
				if slashingParamsQuery != nil {
					rpcQueries = append(rpcQueries, *slashingParamsQuery)
				}
				if stakingParamsQuery != nil {
					rpcQueries = append(rpcQueries, *stakingParamsQuery)
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

func (m *Manager) GetValidatorRankAndTotalStake(chain Chain, address string, rpc *RPC) (uint64, int, float64, float64, QueryInfo, error) {
	allValidators, info, err := rpc.GetAllValidators()
	if err != nil {
		m.Logger.Error().
			Err(err).
			Str("chain", chain.Name).
			Str("address", address).
			Msg("Error querying for validators")
		return 0, 0, 0, 0, info, err
	}

	activeValidators := Filter(allValidators.Validators, func(v Validator) bool {
		return v.Status == "BOND_STATUS_BONDED"
	})

	sort.Slice(activeValidators, func(i, j int) bool {
		return StrToFloat64(activeValidators[i].DelegatorShares) > StrToFloat64(activeValidators[j].DelegatorShares)
	})

	lastValidatorStake := StrToFloat64(activeValidators[len(activeValidators)-1].DelegatorShares)
	var validatorRank uint64 = 0
	var totalStake float64 = 0

	for index, validator := range activeValidators {
		totalStake += StrToFloat64(validator.DelegatorShares)
		if validator.OperatorAddress == address {
			validatorRank = uint64(index) + 1
		}
	}

	return validatorRank, len(activeValidators), totalStake, lastValidatorStake, info, nil
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
