package generators

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"main/pkg/utils"

	"github.com/prometheus/client_golang/prometheus"
)

type SingleValidatorInfoGenerator struct {
	Chains []configPkg.Chain
}

func NewSingleValidatorInfoGenerator(chains []configPkg.Chain) *SingleValidatorInfoGenerator {
	return &SingleValidatorInfoGenerator{Chains: chains}
}

func (g *SingleValidatorInfoGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	dataRaw, ok := state.Get(constants.FetcherNameValidators)
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

	isActiveGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "active",
			Help: "Whether a validator is active (1 if yes, 0 if no)",
		},
		[]string{"chain", "address"},
	)

	commissionGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "commission",
			Help: "Validator current commission",
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
		[]string{"chain", "address"},
	)

	data, _ := dataRaw.(fetchersPkg.ValidatorsData)

	for _, chain := range g.Chains {
		chainValidators, ok := data.Validators[chain.Name]
		if !ok {
			continue
		}

		for _, validatorAddr := range chain.Validators {
			validator, ok := utils.Find(chainValidators.Validators, func(v types.Validator) bool {
				return v.OperatorAddress == validatorAddr.Address
			})

			if !ok {
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

			isActiveGauge.With(prometheus.Labels{
				"chain":   chain.Name,
				"address": validatorAddr.Address,
			}).Set(utils.BoolToFloat64(validator.Active()))

			commissionGauge.With(prometheus.Labels{
				"chain":   chain.Name,
				"address": validatorAddr.Address,
			}).Set(utils.StrToFloat64(validator.Commission.CommissionRates.Rate))

			commissionMaxGauge.With(prometheus.Labels{
				"chain":   chain.Name,
				"address": validatorAddr.Address,
			}).Set(utils.StrToFloat64(validator.Commission.CommissionRates.MaxRate))

			commissionMaxChangeGauge.With(prometheus.Labels{
				"chain":   chain.Name,
				"address": validatorAddr.Address,
			}).Set(utils.StrToFloat64(validator.Commission.CommissionRates.MaxChangeRate))

			delegationsGauge.With(prometheus.Labels{
				"chain":   chain.Name,
				"address": validatorAddr.Address,
			}).Set(utils.StrToFloat64(validator.DelegatorShares))
		}
	}

	return []prometheus.Collector{
		validatorInfoGauge,
		isJailedGauge,
		isActiveGauge,
		commissionGauge,
		commissionMaxGauge,
		commissionMaxChangeGauge,
		delegationsGauge,
	}
}
