package generators

import (
	"main/pkg/config"
	statePkg "main/pkg/state"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
)

func TestIsConsumerGenerator(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	chains := []*config.Chain{
		{
			Name: "chain",
			ConsumerChains: []*config.ConsumerChain{
				{
					Name: "consumer",
				},
			},
		},
	}

	generator := NewIsConsumerGenerator(chains)
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Zero(t, testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain": "chain",
	})))
	assert.InEpsilon(t, float64(1), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain": "consumer",
	})), 0.01)
}
