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

func TestCommissionGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewCommissionGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestCommissionGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameCommission, fetchers.CommissionData{
		Commissions: map[string]map[string][]types.Amount{
			"chain": {
				"validator": []types.Amount{
					{Amount: 1, Denom: "denom"},
				},
			},
		},
	})

	generator := NewCommissionGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(1), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain":   "chain",
		"address": "validator",
		"denom":   "denom",
	})), 0.01)
}
