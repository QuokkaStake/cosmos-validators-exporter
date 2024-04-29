package queriers

import (
	"context"
	"main/pkg/config"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"
	"sort"
	"sync"

	"go.opentelemetry.io/otel/trace"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type ValidatorQuerier struct {
	Logger zerolog.Logger
	Config *config.Config
	Tracer trace.Tracer
}

func NewValidatorQuerier(
	logger *zerolog.Logger,
	config *config.Config,
	tracer trace.Tracer,
) *ValidatorQuerier {
	return &ValidatorQuerier{
		Logger: logger.With().Str("component", "validator_querier").Logger(),
		Config: config,
		Tracer: tracer,
	}
}

func (q *ValidatorQuerier) GetMetrics(ctx context.Context) ([]prometheus.Collector, []*types.QueryInfo) {
	var queryInfos []*types.QueryInfo

	var wg sync.WaitGroup
	var mutex sync.Mutex

	validatorInfoGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_info",
			Help: "Validator info",
		},
		[]string{
			"chain",
			"address",
			"moniker",
			"details",
			"identity",
			"security_contact",
			"website",
		},
	)

	isJailedGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_jailed",
			Help: "Whether a validator is jailed (1 if yes, 0 if no)",
		},
		[]string{"chain", "address"},
	)

	isActiveGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_active",
			Help: "Whether a validator is active (1 if yes, 0 if no)",
		},
		[]string{"chain", "address"},
	)

	commissionGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_commission",
			Help: "Validator current commission",
		},
		[]string{"chain", "address"},
	)

	commissionMaxGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_commission_max",
			Help: "Max commission for validator",
		},
		[]string{"chain", "address"},
	)

	commissionMaxChangeGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_commission_max_change",
			Help: "Max commission change for validator",
		},
		[]string{"chain", "address"},
	)

	delegationsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_total_delegations",
			Help: "Validator delegations (in tokens)",
		},
		[]string{"chain", "address"},
	)

	activeSetSizeGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_active_set_size",
			Help: "Active set size",
		},
		[]string{"chain"},
	)

	validatorRankGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_rank",
			Help: "Rank of a validator compared to other validators on chain.",
		},
		[]string{"chain", "address"},
	)

	validatorsCountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_validators_count",
			Help: "Total active validators count on chain.",
		},
		[]string{"chain", "address"},
	)

	activeSetTokensGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_active_set_tokens",
			Help: "Tokens needed to get into active set (last validators' stake, or 0 if not enough validators)",
		},
		[]string{"chain"},
	)

	totalBondedTokensGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_tokens_bonded_total",
			Help: "Total tokens bonded in chain",
		},
		[]string{"chain"},
	)

	for _, chain := range q.Config.Chains {
		rpc := tendermint.NewRPC(chain, q.Config.Timeout, q.Logger, q.Tracer)

		for _, validator := range chain.Validators {
			wg.Add(1)
			go func(validator string, rpc *tendermint.RPC, chain config.Chain) {
				var (
					validatorInfo       *types.ValidatorResponse
					validatorQueryInfo  *types.QueryInfo
					validatorQueryError error

					allValidators           *types.ValidatorsResponse
					allValidatorsQueryInfo  *types.QueryInfo
					allValidatorsQueryError error

					stakingParams           *types.StakingParamsResponse
					stakingParamsQuery      *types.QueryInfo
					stakingParamsQueryError error

					internalWg sync.WaitGroup
				)

				internalWg.Add(1)

				go func() {
					defer internalWg.Done()
					validatorInfo, validatorQueryInfo, validatorQueryError = rpc.GetValidator(validator, ctx)
					if validatorQueryError != nil {
						q.Logger.Error().
							Err(validatorQueryError).
							Str("chain", chain.Name).
							Str("address", validator).
							Msg("Error querying for validator info")
					}
				}()

				internalWg.Add(1)
				go func() {
					defer internalWg.Done()

					stakingParams, stakingParamsQuery, stakingParamsQueryError = rpc.GetStakingParams(ctx)

					if stakingParamsQueryError != nil {
						q.Logger.Error().
							Err(stakingParamsQueryError).
							Str("chain", chain.Name).
							Str("address", validator).
							Msg("Error querying staking params")
					}
				}()

				internalWg.Add(1)
				go func() {
					allValidators, allValidatorsQueryInfo, allValidatorsQueryError = rpc.GetAllValidators(ctx)

					if allValidatorsQueryError != nil {
						q.Logger.Error().
							Err(stakingParamsQueryError).
							Str("chain", chain.Name).
							Str("address", validator).
							Msg("Error querying all validators")
					}
					internalWg.Done()
				}()

				internalWg.Wait()

				mutex.Lock()
				defer mutex.Unlock()

				if validatorQueryInfo != nil {
					queryInfos = append(queryInfos, validatorQueryInfo)
				}
				if stakingParamsQuery != nil {
					queryInfos = append(queryInfos, stakingParamsQuery)
				}
				if allValidatorsQueryInfo != nil {
					queryInfos = append(queryInfos, allValidatorsQueryInfo)
				}

				// validator request may fail or be disabled, here it's assumed it didn't
				if validatorInfo != nil && validatorInfo.Validator.Description.Moniker != "" {
					validatorInfoGauge.With(prometheus.Labels{
						"chain":            chain.Name,
						"address":          validator,
						"moniker":          validatorInfo.Validator.Description.Moniker,
						"details":          validatorInfo.Validator.Description.Details,
						"identity":         validatorInfo.Validator.Description.Identity,
						"security_contact": validatorInfo.Validator.Description.SecurityContact,
						"website":          validatorInfo.Validator.Description.Website,
					}).Set(1)

					isJailedGauge.With(prometheus.Labels{
						"chain":   chain.Name,
						"address": validator,
					}).Set(utils.BoolToFloat64(validatorInfo.Validator.Jailed))

					isActiveGauge.With(prometheus.Labels{
						"chain":   chain.Name,
						"address": validator,
					}).Set(utils.BoolToFloat64(validatorInfo.Validator.Status == "BOND_STATUS_BONDED"))

					commissionGauge.With(prometheus.Labels{
						"chain":   chain.Name,
						"address": validator,
					}).Set(utils.StrToFloat64(validatorInfo.Validator.Commission.CommissionRates.Rate))

					commissionMaxGauge.With(prometheus.Labels{
						"chain":   chain.Name,
						"address": validator,
					}).Set(utils.StrToFloat64(validatorInfo.Validator.Commission.CommissionRates.MaxRate))

					commissionMaxChangeGauge.With(prometheus.Labels{
						"chain":   chain.Name,
						"address": validator,
					}).Set(utils.StrToFloat64(validatorInfo.Validator.Commission.CommissionRates.MaxChangeRate))

					delegationsGauge.With(prometheus.Labels{
						"chain":   chain.Name,
						"address": validator,
					}).Set(utils.StrToFloat64(validatorInfo.Validator.DelegatorShares))
				}

				if stakingParams != nil {
					maxValidators := int64(stakingParams.StakingParams.MaxValidators)
					if maxValidators >= 0 {
						activeSetSizeGauge.With(prometheus.Labels{
							"chain": chain.Name,
						}).Set(float64(maxValidators))
					}
				}

				if allValidators != nil && len(allValidators.Validators) > 0 {
					activeValidators := utils.Filter(allValidators.Validators, func(v types.Validator) bool {
						return v.Status == "BOND_STATUS_BONDED"
					})

					sort.Slice(activeValidators, func(i, j int) bool {
						return utils.StrToFloat64(activeValidators[i].DelegatorShares) > utils.StrToFloat64(activeValidators[j].DelegatorShares)
					})

					lastValidatorStake := utils.StrToFloat64(activeValidators[len(activeValidators)-1].DelegatorShares)
					var validatorRank uint64 = 0
					var totalStake float64 = 0

					for index, activeValidator := range activeValidators {
						totalStake += utils.StrToFloat64(activeValidator.DelegatorShares)
						if activeValidator.OperatorAddress == validator {
							validatorRank = uint64(index) + 1
						}
					}

					if validatorRank != 0 {
						validatorRankGauge.With(prometheus.Labels{
							"chain":   chain.Name,
							"address": validator,
						}).Set(float64(validatorRank))
					}

					validatorsCountGauge.With(prometheus.Labels{
						"chain":   chain.Name,
						"address": validator,
					}).Set(float64(len(activeValidators)))

					totalBondedTokensGauge.With(prometheus.Labels{
						"chain": chain.Name,
					}).Set(totalStake)

					if stakingParams != nil && len(activeValidators) >= stakingParams.StakingParams.MaxValidators {
						activeSetTokensGauge.With(prometheus.Labels{
							"chain": chain.Name,
						}).Set(lastValidatorStake)
					} else {
						activeSetTokensGauge.With(prometheus.Labels{
							"chain": chain.Name,
						}).Set(0)
					}
				}

				wg.Done()
			}(validator.Address, rpc, chain)
		}
	}

	wg.Wait()

	return []prometheus.Collector{
		validatorInfoGauge,
		isJailedGauge,
		isActiveGauge,
		commissionGauge,
		commissionMaxGauge,
		commissionMaxChangeGauge,
		delegationsGauge,
		activeSetSizeGauge,
		validatorRankGauge,
		validatorsCountGauge,
		activeSetTokensGauge,
		totalBondedTokensGauge,
	}, queryInfos
}

func (q *ValidatorQuerier) Name() string {
	return "validator-querier"
}
