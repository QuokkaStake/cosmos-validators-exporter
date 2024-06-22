package generators

import (
	"main/pkg/config"
	statePkg "main/pkg/state"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
)

func TestDenomCoefficientGenerator(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	chains := []*config.Chain{
		{
			Name:      "chain",
			BaseDenom: "uatom",
			Denoms: config.DenomInfos{
				{Denom: "uatom", DisplayDenom: "atom", DenomCoefficient: 1000000},
			},
			ConsumerChains: []*config.ConsumerChain{
				{
					Name:      "consumer",
					BaseDenom: "uakt",
					Denoms: config.DenomInfos{
						{Denom: "uakt", DisplayDenom: "akt", DenomCoefficient: 1000000},
					},
				},
			},
		},
	}

	generator := NewDenomCoefficientGenerator(chains)
	results := generator.Generate(state)
	assert.Len(t, results, 2)

	coefficientGauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 2, testutil.CollectAndCount(coefficientGauge))
	assert.InEpsilon(t, float64(1000000), 0.01, testutil.ToFloat64(coefficientGauge.With(prometheus.Labels{
		"chain":         "chain",
		"denom":         "uatom",
		"display_denom": "atom",
	})))
	assert.InEpsilon(t, float64(1000000), 0.01, testutil.ToFloat64(coefficientGauge.With(prometheus.Labels{
		"chain":         "consumer",
		"denom":         "uakt",
		"display_denom": "akt",
	})))

	baseDenomGauge, ok := results[1].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 2, testutil.CollectAndCount(baseDenomGauge))
	assert.InEpsilon(t, float64(1), testutil.ToFloat64(baseDenomGauge.With(prometheus.Labels{
		"chain": "chain",
		"denom": "uatom",
	})), 0.01)
	assert.InEpsilon(t, float64(1), testutil.ToFloat64(baseDenomGauge.With(prometheus.Labels{
		"chain": "consumer",
		"denom": "uakt",
	})), 0.01)
}
