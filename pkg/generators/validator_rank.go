package generators

import (
	"github.com/rs/zerolog"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"main/pkg/utils"
	"sort"

	"github.com/prometheus/client_golang/prometheus"
)

type ValidatorRankGenerator struct {
	Chains []configPkg.Chain
	Logger zerolog.Logger
}

func NewValidatorRankGenerator(
	chains []configPkg.Chain,
	logger *zerolog.Logger,
) *ValidatorRankGenerator {
	return &ValidatorRankGenerator{
		Chains: chains,
		Logger: logger.With().Str("component", "validator_rank_generator").Logger(),
	}
}

func (g *ValidatorRankGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	dataRaw, ok := state.Get(constants.FetcherNameValidators)
	if !ok {
		return []prometheus.Collector{}
	}

	validatorRankGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "rank",
			Help: "Rank of a validator compared to other validators on chain.",
		},
		[]string{"chain", "address"},
	)

	data, _ := dataRaw.(fetchersPkg.ValidatorsData)

	for _, chain := range g.Chains {
		chainValidators, ok := data.Validators[chain.Name]
		if !ok {
			g.Logger.Warn().
				Str("chain", chain.Name).
				Msg("Could not find validators list")
			continue
		}

		for _, validatorAddr := range chain.Validators {
			validator, ok := utils.Find(chainValidators.Validators, func(v types.Validator) bool {
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

			activeValidators := utils.Filter(chainValidators.Validators, func(v types.Validator) bool {
				return v.Active()
			})

			sort.Slice(activeValidators, func(i, j int) bool {
				return utils.StrToFloat64(activeValidators[i].DelegatorShares) > utils.StrToFloat64(activeValidators[j].DelegatorShares)
			})

			rank, found := utils.FindIndex(activeValidators, func(v types.Validator) bool {
				return v.OperatorAddress == validator.OperatorAddress
			})

			if found {
				validatorRank := uint64(rank) + 1

				validatorRankGauge.With(prometheus.Labels{
					"chain":   chain.Name,
					"address": validator.OperatorAddress,
				}).Set(float64(validatorRank))
			}
		}
	}

	return []prometheus.Collector{validatorRankGauge}
}
