package generators

import (
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"main/pkg/utils"
	"sort"

	"cosmossdk.io/math"

	"github.com/prometheus/client_golang/prometheus"
)

type ValidatorsInfoGenerator struct {
}

func NewValidatorsInfoGenerator() *ValidatorsInfoGenerator {
	return &ValidatorsInfoGenerator{}
}

func (g *ValidatorsInfoGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	dataRaw, ok := state.Get(constants.FetcherNameValidators)
	if !ok {
		return []prometheus.Collector{}
	}

	consumersDataRaw, ok := state.Get(constants.FetcherNameConsumerValidators)
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
		[]string{"chain"},
	)

	data, _ := dataRaw.(fetchersPkg.ValidatorsData)
	consumersData, _ := consumersDataRaw.(fetchersPkg.ConsumerValidatorsData)

	for chain, validators := range data.Validators {
		activeValidators := utils.Filter(validators.Validators, func(v types.Validator) bool {
			return v.Active()
		})

		sort.Slice(activeValidators, func(i, j int) bool {
			return activeValidators[i].DelegatorShares.GT(activeValidators[j].DelegatorShares)
		})

		totalStake := math.LegacyNewDec(0)

		for _, activeValidator := range activeValidators {
			totalStake = totalStake.Add(activeValidator.DelegatorShares)
		}

		validatorsCountGauge.With(prometheus.Labels{
			"chain": chain,
		}).Set(float64(len(activeValidators)))

		totalBondedTokensGauge.With(prometheus.Labels{
			"chain": chain,
		}).Set(totalStake.MustFloat64())
	}

	for chain, validators := range consumersData.Validators {
		validatorsCountGauge.With(prometheus.Labels{
			"chain": chain,
		}).Set(float64(len(validators.Validators)))
	}

	return []prometheus.Collector{validatorsCountGauge, totalBondedTokensGauge}
}
