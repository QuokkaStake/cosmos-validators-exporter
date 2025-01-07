package generators

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"main/pkg/utils"

	"github.com/rs/zerolog"

	"github.com/prometheus/client_golang/prometheus"
)

type SingleValidatorInfoGenerator struct {
	Chains []*configPkg.Chain
	Logger zerolog.Logger
}

func NewSingleValidatorInfoGenerator(
	chains []*configPkg.Chain,
	logger *zerolog.Logger,
) *SingleValidatorInfoGenerator {
	return &SingleValidatorInfoGenerator{
		Chains: chains,
		Logger: logger.With().Str("component", "single_validator_info_generator").Logger(),
	}
}

func (g *SingleValidatorInfoGenerator) Generate(state statePkg.State) []prometheus.Collector {
	data, ok := statePkg.StateGet[fetchersPkg.ValidatorsData](state, constants.FetcherNameValidators)
	if !ok {
		return []prometheus.Collector{}
	}

	validatorInfoGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "info",
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
			Name: constants.MetricsPrefix + "jailed",
			Help: "Whether a validator is jailed (1 if yes, 0 if no)",
		},
		[]string{"chain", "address"},
	)

	commissionMaxGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "commission_max",
			Help: "Max commission for validator",
		},
		[]string{"chain", "address"},
	)

	commissionMaxChangeGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "commission_max_change",
			Help: "Max commission change for validator",
		},
		[]string{"chain", "address"},
	)

	delegationsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "total_delegations",
			Help: "Validator delegations (in tokens)",
		},
		[]string{"chain", "address", "denom"},
	)

	for _, chain := range g.Chains {
		chainValidators, ok := data.Validators[chain.Name]
		if !ok {
			g.Logger.Warn().
				Str("chain", chain.Name).
				Msg("Could not find validators list")
			continue
		}

		for _, validatorAddr := range chain.Validators {
			compare := func(v types.Validator) bool {
				equal, err := utils.CompareTwoBech32(v.OperatorAddress, validatorAddr.Address)
				if err != nil {
					g.Logger.Error().
						Err(err).
						Str("chain", chain.Name).
						Str("validator", validatorAddr.Address).
						Msg("Error comparing two validators' bech32 addresses")
					return false
				}

				return equal
			}

			validator, ok := utils.Find(chainValidators.Validators, compare)

			if !ok {
				g.Logger.Warn().
					Str("chain", chain.Name).
					Str("validator", validatorAddr.Address).
					Msg("Could not find validator")
				continue
			}

			validatorInfoGauge.With(prometheus.Labels{
				"chain":            chain.Name,
				"address":          validatorAddr.Address,
				"moniker":          validator.Description.Moniker,
				"details":          validator.Description.Details,
				"identity":         validator.Description.Identity,
				"security_contact": validator.Description.SecurityContact,
				"website":          validator.Description.Website,
			}).Set(1)

			isJailedGauge.With(prometheus.Labels{
				"chain":   chain.Name,
				"address": validatorAddr.Address,
			}).Set(utils.BoolToFloat64(validator.Jailed))

			commissionMaxGauge.With(prometheus.Labels{
				"chain":   chain.Name,
				"address": validatorAddr.Address,
			}).Set(validator.Commission.CommissionRates.MaxRate.MustFloat64())

			commissionMaxChangeGauge.With(prometheus.Labels{
				"chain":   chain.Name,
				"address": validatorAddr.Address,
			}).Set(validator.Commission.CommissionRates.MaxChangeRate.MustFloat64())

			delegationsAmount := chain.Denoms.Convert(&types.Amount{
				Amount: validator.DelegatorShares.MustFloat64(),
				Denom:  chain.BaseDenom,
			})

			if delegationsAmount == nil {
				continue
			}

			delegationsGauge.With(prometheus.Labels{
				"chain":   chain.Name,
				"address": validatorAddr.Address,
				"denom":   delegationsAmount.Denom,
			}).Set(delegationsAmount.Amount)
		}
	}

	return []prometheus.Collector{
		validatorInfoGauge,
		isJailedGauge,
		commissionMaxGauge,
		commissionMaxChangeGauge,
		delegationsGauge,
	}
}
