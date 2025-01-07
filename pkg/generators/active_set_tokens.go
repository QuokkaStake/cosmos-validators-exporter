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
	Chains []*configPkg.Chain
}

func NewActiveSetTokensGenerator(chains []*configPkg.Chain) *ActiveSetTokensGenerator {
	return &ActiveSetTokensGenerator{Chains: chains}
}

func (g *ActiveSetTokensGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	validators, ok := statePkg.StateGet[fetchersPkg.ValidatorsData](state, constants.FetcherNameValidators)
	if !ok {
		return []prometheus.Collector{}
	}

	stakingParams, ok := statePkg.StateGet[fetchersPkg.StakingParamsData](state, constants.FetcherNameStakingParams)
	if !ok {
		return []prometheus.Collector{}
	}

	activeSetTokensGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "active_set_tokens",
			Help: "Tokens needed to get into active set (last validators' stake, or 0 if not enough validators)",
		},
		[]string{"chain", "denom"},
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
			return v.Active()
		})

		sort.Slice(activeValidators, func(i, j int) bool {
			return activeValidators[i].DelegatorShares.GT(activeValidators[j].DelegatorShares)
		})

		lastValidatorStake := activeValidators[len(activeValidators)-1].DelegatorShares.MustFloat64()
		lastValidatorAmount := chain.Denoms.Convert(&types.Amount{
			Amount: lastValidatorStake,
			Denom:  chain.BaseDenom,
		})

		if lastValidatorAmount == nil {
			continue
		}

		if chainStakingParams != nil && len(activeValidators) >= chainStakingParams.StakingParams.MaxValidators {
			activeSetTokensGauge.With(prometheus.Labels{
				"chain": chain.Name,
				"denom": lastValidatorAmount.Denom,
			}).Set(lastValidatorAmount.Amount)
		} else {
			activeSetTokensGauge.With(prometheus.Labels{
				"chain": chain.Name,
				"denom": lastValidatorAmount.Denom,
			}).Set(0)
		}
	}

	return []prometheus.Collector{activeSetTokensGauge}
}
