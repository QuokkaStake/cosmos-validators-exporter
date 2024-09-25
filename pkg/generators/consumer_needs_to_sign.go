package generators

import (
	"main/pkg/config"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/utils"

	"github.com/prometheus/client_golang/prometheus"
)

type ConsumerNeedsToSignGenerator struct {
	Chains []*config.Chain
}

func NewConsumerNeedsToSignGenerator(chains []*config.Chain) *ConsumerNeedsToSignGenerator {
	return &ConsumerNeedsToSignGenerator{Chains: chains}
}

func (g *ConsumerNeedsToSignGenerator) Generate(state *statePkg.State) []prometheus.Collector {
	allValidatorsConsumersRaw, ok := state.Get(constants.FetcherNameValidatorConsumers)
	if !ok {
		return []prometheus.Collector{}
	}

	consumerInfosRaw, ok := state.Get(constants.FetcherNameConsumerInfo)
	if !ok {
		return []prometheus.Collector{}
	}

	allValidatorsConsumers, _ := allValidatorsConsumersRaw.(fetchersPkg.ValidatorConsumersData)
	consumerInfos, _ := consumerInfosRaw.(fetchersPkg.ConsumerInfoData)

	needsToSignGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "consumer_needs_to_sign",
			Help: "Top-N percent threshold for consumer chains.",
		},
		[]string{
			"consumer_id",
			"provider",
			"address",
		},
	)

	for _, chain := range g.Chains {
		chainInfos, ok := allValidatorsConsumers.Infos[chain.Name]
		if !ok {
			continue
		}

		chainConsumers, ok := consumerInfos.Info[chain.Name]
		if !ok {
			continue
		}

		for _, validator := range chain.Validators {
			validatorConsumers, ok := chainInfos[validator.Address]
			if !ok {
				continue
			}

			for _, consumer := range chainConsumers {
				_, needsToSign := validatorConsumers[consumer.ConsumerID]

				needsToSignGauge.With(prometheus.Labels{
					"consumer_id": consumer.ConsumerID,
					"provider":    chain.Name,
					"address":     validator.Address,
				}).Set(utils.BoolToFloat64(needsToSign))
			}
		}
	}

	return []prometheus.Collector{needsToSignGauge}
}
