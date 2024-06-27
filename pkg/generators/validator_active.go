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

type ValidatorActiveGenerator struct {
	Chains []*configPkg.Chain
	Logger zerolog.Logger
}

func NewValidatorActiveGenerator(
	chains []*configPkg.Chain,
	logger *zerolog.Logger,
) *ValidatorActiveGenerator {
	return &ValidatorActiveGenerator{
		Chains: chains,
		Logger: logger.With().Str("component", "validator_active_generator").Logger(),
	}
}

func (g *ValidatorActiveGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	dataRaw, ok := state.Get(constants.FetcherNameValidators)
	if !ok {
		return []prometheus.Collector{}
	}

	consumersDataRaw, ok := state.Get(constants.FetcherNameConsumerValidators)
	if !ok {
		return []prometheus.Collector{}
	}

	isActiveGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "active",
			Help: "Whether a validator is active (1 if yes, 0 if no)",
		},
		[]string{"chain", "address"},
	)

	data, _ := dataRaw.(fetchersPkg.ValidatorsData)
	consumersData, _ := consumersDataRaw.(fetchersPkg.ConsumerValidatorsData)

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

			isActiveGauge.With(prometheus.Labels{
				"chain":   chain.Name,
				"address": validatorAddr.Address,
			}).Set(utils.BoolToFloat64(validator.Active()))

			if validatorAddr.ConsensusAddress == "" {
				continue
			}

			for _, consumer := range chain.ConsumerChains {
				consumerValidators, ok := consumersData.Validators[consumer.Name]
				if !ok {
					continue
				}

				_, isActive := utils.Find(consumerValidators.Validators, func(v types.ConsumerValidator) bool {
					return v.ProviderAddress == validatorAddr.ConsensusAddress
				})

				isActiveGauge.With(prometheus.Labels{
					"chain":   consumer.Name,
					"address": validatorAddr.Address,
				}).Set(utils.BoolToFloat64(isActive))
			}
		}
	}

	return []prometheus.Collector{
		isActiveGauge,
	}
}
