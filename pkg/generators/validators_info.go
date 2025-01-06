package generators

import (
	"main/pkg/config"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	"main/pkg/types"
	"main/pkg/utils"

	"cosmossdk.io/math"

	"github.com/prometheus/client_golang/prometheus"
)

type ValidatorsInfoGenerator struct {
	Chains []*config.Chain
}

func NewValidatorsInfoGenerator(chains []*config.Chain) *ValidatorsInfoGenerator {
	return &ValidatorsInfoGenerator{Chains: chains}
}

func (g *ValidatorsInfoGenerator) Generate(state fetchersPkg.State) []prometheus.Collector {
	data, ok := fetchersPkg.StateGet[fetchersPkg.ValidatorsData](state, constants.FetcherNameValidators)
	if !ok {
		return []prometheus.Collector{}
	}

	consumersData, ok := fetchersPkg.StateGet[fetchersPkg.ConsumerValidatorsData](state, constants.FetcherNameConsumerValidators)
	if !ok {
		return []prometheus.Collector{}
	}

	validatorsCountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "validators_count",
			Help: "Total active validators count on chain.",
		},
		[]string{"chain"},
	)

	totalBondedTokensGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "tokens_bonded_total",
			Help: "Total tokens bonded in chain",
		},
		[]string{"chain", "denom"},
	)

	for _, chain := range g.Chains {
		validators, ok := data.Validators[chain.Name]
		if !ok {
			continue
		}

		activeValidators := utils.Filter(validators.Validators, func(v types.Validator) bool {
			return v.Active()
		})

		totalStake := math.LegacyNewDec(0)

		for _, activeValidator := range activeValidators {
			totalStake = totalStake.Add(activeValidator.DelegatorShares)
		}

		validatorsCountGauge.With(prometheus.Labels{
			"chain": chain.Name,
		}).Set(float64(len(activeValidators)))

		totalBondedAmount := &types.Amount{
			Amount: totalStake.MustFloat64(),
			Denom:  chain.BaseDenom,
		}
		totalBondedAmountConverted := chain.Denoms.Convert(totalBondedAmount)
		if totalBondedAmountConverted == nil {
			continue
		}

		totalBondedTokensGauge.With(prometheus.Labels{
			"chain": chain.Name,
			"denom": totalBondedAmountConverted.Denom,
		}).Set(totalBondedAmountConverted.Amount)
	}

	for chain, validators := range consumersData.Validators {
		validatorsCountGauge.With(prometheus.Labels{
			"chain": chain,
		}).Set(float64(len(validators.Validators)))
	}

	return []prometheus.Collector{validatorsCountGauge, totalBondedTokensGauge}
}
