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

	"github.com/stretchr/testify/assert"
)

func TestCommissionGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewCommissionGenerator([]*config.Chain{})
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
					{Amount: 100000, Denom: "uatom"},
					{Amount: 200000, Denom: "ustake"},
				},
			},
		},
	})

	chains := []*config.Chain{
		{
			Name:       "chain",
			Validators: []config.Validator{{Address: "validator"}, {Address: "validator2"}},
			Denoms: config.DenomInfos{
				{Denom: "uatom", DisplayDenom: "atom", DenomExponent: 6},
			},
		},
		{
			Name:       "chain2",
			Validators: []config.Validator{{Address: "validator"}},
		},
	}
	generator := NewCommissionGenerator(chains)
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, 0.1, testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain":   "chain",
		"address": "validator",
		"denom":   "atom",
	})), 0.01)
	assert.InEpsilon(t, float64(200000), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain":   "chain",
		"address": "validator",
		"denom":   "ustake",
	})), 0.01)
}
