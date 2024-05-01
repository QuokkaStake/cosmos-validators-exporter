package generators

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"main/pkg/utils"
	"math/big"
)

type ConsumerNeedsToSignGenerator struct {
	Chains []configPkg.Chain
	Logger zerolog.Logger
}

func NewConsumerNeedsToSignGenerator(
	chains []configPkg.Chain,
	logger *zerolog.Logger,
) *ConsumerNeedsToSignGenerator {
	return &ConsumerNeedsToSignGenerator{
		Chains: chains,
		Logger: logger.With().Str("component", "consumer_needs_to_sign").Logger(),
	}
}

func (g *ConsumerNeedsToSignGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	validatorsRaw, ok := state.Get(constants.FetcherNameValidators)
	if !ok {
		return []prometheus.Collector{}
	}

	thresholdRaw, ok := state.Get(constants.FetcherNameSoftOptOutThreshold)
	if !ok {
		return []prometheus.Collector{}
	}

	validators, _ := validatorsRaw.(fetchersPkg.ValidatorsData)
	threshold, _ := thresholdRaw.(fetchersPkg.SoftOptOutThresholdData)

	minStakeGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "minimal_stake_to_require_signing",
			Help: "Minimal stake for validator to be required to sign blocks",
		},
		[]string{"chain"},
	)

	needsToSignGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "needs_to_sign",
			Help: "Whether a validator is required to sign blocks (1 if yes, 0 if no)",
		},
		[]string{"chain", "validator"},
	)

	for _, chain := range g.Chains {
		if !chain.IsConsumer() {
			continue
		}

		chainValidators, ok := validators.Validators[chain.Name]
		if !ok {
			continue
		}

		chainSoftOptOutThreshold, ok := threshold.Thresholds[chain.Name]
		if !ok {
			continue
		}

		// sort ascending
		activeValidators := utils.Filter(chainValidators.Validators, func(v types.Validator) bool {
			return v.Active()
		})

		// get total VP
		totalVP := big.NewFloat(0)

		for _, validator := range activeValidators {
			validatorVP, _, err := big.ParseFloat(validator.Tokens, 10, 10, big.ToZero)
			if err != nil {
				g.Logger.Error().
					Err(err).
					Str("chain", chain.Name).
					Str("validator", validator.OperatorAddress).
					Msg("Error parsing validator VP")
				continue
			}
			totalVP.Add(totalVP, validatorVP)
		}

		thresholdVP := big.NewFloat(0)

		// calculate min stake for threshold
		for _, validator := range activeValidators {
			validatorVP, _, err := big.ParseFloat(validator.Tokens, 10, 10, big.ToZero)
			if err != nil {
				g.Logger.Error().
					Err(err).
					Str("chain", chain.Name).
					Str("validator", validator.OperatorAddress).
					Msg("Error parsing validator VP")
				continue
			}

			thresholdVP = big.NewFloat(0).Add(thresholdVP, validatorVP)
			thresholdPercent := big.NewFloat(0).Quo(thresholdVP, totalVP)

			if thresholdPercent.Cmp(big.NewFloat(chainSoftOptOutThreshold)) > 0 {
				thresholdVPFloat, _ := thresholdVP.Float64()

				minStakeGauge.
					With(prometheus.Labels{"chain": chain.Name}).
					Set(thresholdVPFloat)

				break
			}
		}

		for _, validatorAddr := range chain.Validators {
			validator, ok := utils.Find(activeValidators, func(v types.Validator) bool {
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
			})

			if !ok {
				g.Logger.Warn().
					Str("chain", chain.Name).
					Str("validator", validatorAddr.Address).
					Msg("Could not find validator")
				continue
			}

			validatorVP, _, err := big.ParseFloat(validator.Tokens, 10, 10, big.ToZero)
			if err != nil {
				g.Logger.Error().
					Err(err).
					Str("chain", chain.Name).
					Str("validator", validatorAddr.Address).
					Msg("Error parsing validator VP")
				continue
			}

			compare := validatorVP.Cmp(thresholdVP)

			needsToSignGauge.
				With(prometheus.Labels{
					"chain":     chain.Name,
					"validator": validatorAddr.Address,
				}).
				Set(utils.BoolToFloat64(compare >= 0))
		}

	}

	return []prometheus.Collector{minStakeGauge, needsToSignGauge}
}
