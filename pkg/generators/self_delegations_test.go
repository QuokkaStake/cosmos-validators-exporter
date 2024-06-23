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

func TestSelfDelegationGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewSelfDelegationGenerator([]*config.Chain{})
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestSelfDelegationGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameSelfDelegation, fetchers.SelfDelegationData{
		Delegations: map[string]map[string]*types.Amount{
			"chain": {
				"validator":  {Amount: 100000, Denom: "uatom"},
				"validator2": {Amount: 200000, Denom: "ustake"},
			},
		},
	})

	chains := []*config.Chain{
		{
			Name: "chain",
			Validators: []config.Validator{
				{Address: "validator"},
				{Address: "validator2"},
				{Address: "validator3"},
			},
			Denoms: config.DenomInfos{
				{Denom: "uatom", DisplayDenom: "atom", DenomCoefficient: 1000000},
			},
		},
		{
			Name:       "chain2",
			Validators: []config.Validator{{Address: "validator"}},
		},
	}
	generator := NewSelfDelegationGenerator(chains)
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
		"address": "validator2",
		"denom":   "ustake",
	})), 0.01)
}
