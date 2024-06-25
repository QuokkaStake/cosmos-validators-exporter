package generators

import (
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"cosmossdk.io/math"

	"github.com/stretchr/testify/assert"
)

func TestConsumerInfoGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewConsumerInfoGenerator([]*config.Chain{})
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestConsumerInfoGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameConsumerInfo, fetchers.ConsumerInfoData{
		Info: map[string]map[string]types.ConsumerChainInfo{
			"provider": {
				"chain-id": types.ConsumerChainInfo{
					ChainID: "chain-id", TopN: 1, MinPowerInTopN: math.NewInt(100),
				},
			},
		},
	})

	chains := []*config.Chain{{
		Name: "provider",
		ConsumerChains: []*config.ConsumerChain{{
			Name:    "consumer",
			ChainID: "chain-id",
		}, {
			Name:    "otherconsumer",
			ChainID: "otherchainid",
		}},
	}, {Name: "otherprovider"}}

	generator := NewConsumerInfoGenerator(chains)
	results := generator.Generate(state)
	assert.Len(t, results, 2)

	topNGauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, 0.01, testutil.ToFloat64(topNGauge.With(prometheus.Labels{
		"chain":    "consumer",
		"chain_id": "chain-id",
		"provider": "provider",
	})), 0.01)

	minPowerGauge, ok := results[1].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(100), testutil.ToFloat64(minPowerGauge.With(prometheus.Labels{
		"chain":    "consumer",
		"chain_id": "chain-id",
		"provider": "provider",
	})), 0.01)
}
