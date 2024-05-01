package generators

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"main/pkg/utils"
	"sort"

	"github.com/prometheus/client_golang/prometheus"
)

type ActiveSetTokensGenerator struct {
	Chains []configPkg.Chain
}

func NewActiveSetTokensGenerator(chains []configPkg.Chain) *ActiveSetTokensGenerator {
	return &ActiveSetTokensGenerator{Chains: chains}
}

func (g *ActiveSetTokensGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	validatorsRaw, ok := state.Get(constants.FetcherNameValidators)
	if !ok {
		return []prometheus.Collector{}
	}

	stakingParamsRaw, ok := state.Get(constants.FetcherNameStakingParams)
	if !ok {
		return []prometheus.Collector{}
	}

	validators, _ := validatorsRaw.(fetchersPkg.ValidatorsData)
	stakingParams, _ := stakingParamsRaw.(fetchersPkg.StakingParamsData)

	activeSetTokensGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "active_set_tokens",
			Help: "Tokens needed to get into active set (last validators' stake, or 0 if not enough validators)",
		},
		[]string{"chain"},
	)

	for _, chain := range g.Chains {
		chainValidators, ok := validators.Validators[chain.Name]
		if !ok {
			continue
		}

		chainStakingParams, ok := stakingParams.Params[chain.Name]
		if !ok {
			continue
		}

		activeValidators := utils.Filter(chainValidators.Validators, func(v types.Validator) bool {
			return v.Status == "BOND_STATUS_BONDED"
		})

		sort.Slice(activeValidators, func(i, j int) bool {
			return utils.StrToFloat64(activeValidators[i].DelegatorShares) > utils.StrToFloat64(activeValidators[j].DelegatorShares)
		})

		lastValidatorStake := utils.StrToFloat64(activeValidators[len(activeValidators)-1].DelegatorShares)

		if chainStakingParams != nil && len(activeValidators) >= chainStakingParams.StakingParams.MaxValidators {
			activeSetTokensGauge.With(prometheus.Labels{
				"chain": chain.Name,
			}).Set(lastValidatorStake)
		} else {
			activeSetTokensGauge.With(prometheus.Labels{
				"chain": chain.Name,
			}).Set(0)
		}
	}

	return []prometheus.Collector{activeSetTokensGauge}
}
