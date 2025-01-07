package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
)

func TestDelegationsGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewDelegationsGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestDelegationsGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameDelegations, fetchers.DelegationsData{
		Delegations: map[string]map[string]uint64{
			"chain": {
				"validator": 100,
			},
		},
	})

	generator := NewDelegationsGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(100), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain":   "chain",
		"address": "validator",
	})), 0.01)
}
