package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
)

func TestBalanceGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewBalanceGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestBalanceGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameBalance, fetchers.BalanceData{
		Balances: map[string]map[string][]types.Amount{
			"chain": {
				"validator": {
					{Amount: 100, Denom: "uatom"},
					{Amount: 200, Denom: "ustake"},
				},
			},
		},
	})

	generator := NewBalanceGenerator()
	results := generator.Generate(state)
	assert.Len(t, results, 1)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(100), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain":   "chain",
		"address": "validator",
		"denom":   "uatom",
	})), 0.01)
	assert.InEpsilon(t, float64(200), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain":   "chain",
		"address": "validator",
		"denom":   "ustake",
	})), 0.01)
}
