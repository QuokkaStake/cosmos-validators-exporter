package generators

import (
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
	generator := NewConsumerInfoGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestConsumerInfoGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameConsumerInfo, fetchers.ConsumerInfoData{
		Info: map[string]*types.ConsumerInfoResponse{
			"chain": {
				Chains: []types.ConsumerChainInfo{
					{ChainID: "chain-id", TopN: 1, MinPowerInTopN: math.NewInt(100)},
				},
			},
		},
	})

	generator := NewConsumerInfoGenerator()
	results := generator.Generate(state)
	assert.Len(t, results, 2)

	topNGauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, 0.01, testutil.ToFloat64(topNGauge.With(prometheus.Labels{
		"chain":    "chain",
		"chain_id": "chain-id",
	})), 0.01)

	minPowerGauge, ok := results[1].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(100), testutil.ToFloat64(minPowerGauge.With(prometheus.Labels{
		"chain":    "chain",
		"chain_id": "chain-id",
	})), 0.01)
}
