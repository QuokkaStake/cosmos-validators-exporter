package manager

import (
	"sort"
	"sync"

	"main/pkg/config"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"

	"github.com/rs/zerolog"
)

type Manager struct {
	Config *config.Config
	Logger zerolog.Logger
}

func NewManager(config *config.Config, logger *zerolog.Logger) *Manager {
	return &Manager{
		Config: config,
		Logger: logger.With().Str("component", "manager").Logger(),
	}
}

func (m *Manager) GetAllValidators() []types.ValidatorQuery {
	length := 0
	for _, chain := range m.Config.Chains {
		for range chain.Validators {
			length++
		}
	}

	validators := make([]types.ValidatorQuery, length)

	var wg sync.WaitGroup
	wg.Add(length)

	index := 0

	for _, chain := range m.Config.Chains {
		rpc := tendermint.NewRPC(chain, m.Config.Timeout, m.Logger)

		for _, address := range chain.Validators {
			go func(address string, chain config.Chain, index int) {
				defer wg.Done()

				var internalWg sync.WaitGroup

				var (
					info                 *types.ValidatorResponse
					validatorQueryInfo   types.QueryInfo
					validatorQueryError  error
					rank                 uint64
					totalValidators      int
					totalStake           float64
					lastValidatorStake   float64
					validatorsQueryInfo  types.QueryInfo
					validatorsQueryError error

					signingInfo           *types.SigningInfoResponse
					signingInfoQuery      *types.QueryInfo
					signingInfoQueryError error

					stakingParams           *types.StakingParamsResponse
					stakingParamsQuery      *types.QueryInfo
					stakingParamsQueryError error

					validatorInfo types.ValidatorInfo
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
					rank, totalValidators, totalStake, lastValidatorStake, validatorsQueryInfo, validatorsQueryError = m.GetValidatorRankAndTotalStake(chain, address, rpc)
					internalWg.Done()
				}()

				internalWg.Add(1)
				go func() {
					stakingParams, stakingParamsQuery, stakingParamsQueryError = rpc.GetStakingParams()
					internalWg.Done()
				}()

				internalWg.Wait()

				if validatorQueryError != nil {
					m.Logger.Error().
						Err(validatorQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying validator")
					validatorInfo = types.ValidatorInfo{}
				} else {
					validatorInfo = types.NewValidatorInfo(info.Validator)
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
					if stakingParams != nil && totalValidators >= stakingParams.StakingParams.MaxValidators {
						validatorInfo.LastValidatorStake = lastValidatorStake
					}
				}

				if signingInfoQueryError != nil {
					m.Logger.Error().
						Err(signingInfoQueryError).
						Str("chain", chain.Name).
						Str("address", address).
						Msg("Error querying validator signing info")
				} else if signingInfo != nil && signingInfo.ValSigningInfo.Address != "" {
					validatorInfo.MissedBlocksCount = utils.StrToInt64(signingInfo.ValSigningInfo.MissedBlocksCounter)
					validatorInfo.IsTombstoned = signingInfo.ValSigningInfo.Tombstoned
					validatorInfo.JailedUntil = signingInfo.ValSigningInfo.JailedUntil
					validatorInfo.StartHeight = utils.StrToInt64(signingInfo.ValSigningInfo.StartHeight)
					validatorInfo.IndexOffset = utils.StrToInt64(signingInfo.ValSigningInfo.IndexOffset)
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

				rpcQueries := []types.QueryInfo{
					validatorQueryInfo,
					validatorsQueryInfo,
				}

				if signingInfoQuery != nil {
					rpcQueries = append(rpcQueries, *signingInfoQuery)
				}
				if stakingParamsQuery != nil {
					rpcQueries = append(rpcQueries, *stakingParamsQuery)
				}

				query := types.ValidatorQuery{
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

func (m *Manager) GetValidatorRankAndTotalStake(chain config.Chain, address string, rpc *tendermint.RPC) (uint64, int, float64, float64, types.QueryInfo, error) {
	allValidators, info, err := rpc.GetAllValidators()
	if err != nil {
		m.Logger.Error().
			Err(err).
			Str("chain", chain.Name).
			Str("address", address).
			Msg("Error querying for validators")
		return 0, 0, 0, 0, info, err
	}

	activeValidators := utils.Filter(allValidators.Validators, func(v types.Validator) bool {
		return v.Status == "BOND_STATUS_BONDED"
	})

	sort.Slice(activeValidators, func(i, j int) bool {
		return utils.StrToFloat64(activeValidators[i].DelegatorShares) > utils.StrToFloat64(activeValidators[j].DelegatorShares)
	})

	lastValidatorStake := utils.StrToFloat64(activeValidators[len(activeValidators)-1].DelegatorShares)
	var validatorRank uint64 = 0
	var totalStake float64 = 0

	for index, validator := range activeValidators {
		totalStake += utils.StrToFloat64(validator.DelegatorShares)
		if validator.OperatorAddress == address {
			validatorRank = uint64(index) + 1
		}
	}

	return validatorRank, len(activeValidators), totalStake, lastValidatorStake, info, nil
}
