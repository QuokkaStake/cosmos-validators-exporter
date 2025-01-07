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

	state := statePkg.State{}
	generator := NewConsumerInfoGenerator([]*config.Chain{})
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestConsumerInfoGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.State{}
	state.Set(constants.FetcherNameConsumerInfo, fetchers.ConsumerInfoData{
		Info: map[string]map[string]types.ConsumerChainInfo{
			"provider": {
				"0": types.ConsumerChainInfo{
					ChainID:           "chain-id",
					ConsumerID:        "0",
					AllowInactiveVals: false,
					Phase:             "CONSUMER_PHASE_LAUNCHED",
					TopN:              1,
					MinPowerInTopN:    math.NewInt(100),
				},
			},
		},
	})

	chains := []*config.Chain{{
		Name: "provider",
		ConsumerChains: []*config.ConsumerChain{{
			Name:       "consumer",
			ConsumerID: "0",
		}, {
			Name:       "otherconsumer",
			ConsumerID: "1",
		}},
	}, {Name: "otherprovider"}}

	generator := NewConsumerInfoGenerator(chains)
	results := generator.Generate(state)
	assert.Len(t, results, 3)

	consumerInfoGauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(1), testutil.ToFloat64(consumerInfoGauge.With(prometheus.Labels{
		"chain_id":            "chain-id",
		"consumer_id":         "0",
		"provider":            "provider",
		"phase":               "CONSUMER_PHASE_LAUNCHED",
		"allow_inactive_vals": "0",
	})), 0.01)

	topNGauge, ok := results[1].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, 0.01, testutil.ToFloat64(topNGauge.With(prometheus.Labels{
		"consumer_id": "0",
		"provider":    "provider",
	})), 0.01)

	minPowerGauge, ok := results[2].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(100), testutil.ToFloat64(minPowerGauge.With(prometheus.Labels{
		"consumer_id": "0",
		"provider":    "provider",
	})), 0.01)
}
