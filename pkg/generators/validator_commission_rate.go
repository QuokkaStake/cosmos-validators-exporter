package generators

import (
	"main/pkg/config"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"main/pkg/utils"

	"github.com/rs/zerolog"

	"github.com/prometheus/client_golang/prometheus"
)

type ValidatorCommissionRateGenerator struct {
	Chains []*config.Chain
	Logger zerolog.Logger
}

func NewValidatorCommissionRateGenerator(
	chains []*config.Chain,
	logger *zerolog.Logger,
) *ValidatorCommissionRateGenerator {
	return &ValidatorCommissionRateGenerator{
		Chains: chains,
		Logger: logger.With().Str("component", "commission_rate_generator").Logger(),
	}
}

func (g *ValidatorCommissionRateGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	consumerCommissionsRaw, ok := state.Get(constants.FetcherNameConsumerCommission)
	if !ok {
		return []prometheus.Collector{}
	}

	validatorsRaw, ok := state.Get(constants.FetcherNameValidators)
	if !ok {
		return []prometheus.Collector{}
	}

	commissionGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "commission",
			Help: "Validator current commission",
		},
		[]string{"chain", "address"},
	)

	consumerCommission, _ := consumerCommissionsRaw.(fetchersPkg.ConsumerCommissionData)
	validators, _ := validatorsRaw.(fetchersPkg.ValidatorsData)

	for _, chain := range g.Chains {
		chainValidators, ok := validators.Validators[chain.Name]
		if !ok {
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
			} else {
				commissionGauge.With(prometheus.Labels{
					"chain":   chain.Name,
					"address": validatorAddr.Address,
				}).Set(validator.Commission.CommissionRates.Rate.MustFloat64())
			}

			for _, consumer := range chain.ConsumerChains {
				consumerValidators, ok := consumerCommission.Commissions[consumer.Name]
				if !ok {
					continue
				}

				consumerValidatorRate, ok := consumerValidators[validator.OperatorAddress]
				if !ok {
					continue
				}

				commissionGauge.With(prometheus.Labels{
					"chain":   consumer.Name,
					"address": validatorAddr.Address,
				}).Set(consumerValidatorRate.Rate.MustFloat64())
			}
		}
	}

	return []prometheus.Collector{commissionGauge}
}
