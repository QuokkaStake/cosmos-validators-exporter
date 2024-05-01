package generators

import (
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type RewardsGenerator struct {
}

func NewRewardsGenerator() *RewardsGenerator {
	return &RewardsGenerator{}
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

	for chain, rewards := range data.Rewards {
		for validator, validatorReward := range rewards {
			for _, balance := range validatorReward {
				selfDelegationRewardsTokens.With(prometheus.Labels{
					"chain":   chain,
					"address": validator,
					"denom":   balance.Denom,
				}).Set(balance.Amount)
			}
		}
	}

	return []prometheus.Collector{selfDelegationRewardsTokens}
}
