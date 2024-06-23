package generators

import (
	"main/pkg/config"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type RewardsGenerator struct {
	Chains []*config.Chain
}

func NewRewardsGenerator(chains []*config.Chain) *RewardsGenerator {
	return &RewardsGenerator{Chains: chains}
}

func (g *RewardsGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	dataRaw, ok := state.Get(constants.FetcherNameRewards)
	if !ok {
		return []prometheus.Collector{}
	}

	selfDelegationRewardsTokens := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "self_delegation_rewards",
			Help: "Validator's self-delegation rewards (in tokens)",
		},
		[]string{"chain", "address", "denom"},
	)

	data, _ := dataRaw.(fetchersPkg.RewardsData)

	for _, chain := range g.Chains {
		chainRewards, ok := data.Rewards[chain.Name]
		if !ok {
			continue
		}

		for _, validator := range chain.Validators {
			validatorRewards, ok := chainRewards[validator.Address]
			if !ok {
				continue
			}

			for _, balance := range validatorRewards {
				amountConverted := chain.Denoms.Convert(&balance)
				selfDelegationRewardsTokens.With(prometheus.Labels{
					"chain":   chain.Name,
					"address": validator.Address,
					"denom":   amountConverted.Denom,
				}).Set(amountConverted.Amount)
			}
		}
	}

	return []prometheus.Collector{selfDelegationRewardsTokens}
}
