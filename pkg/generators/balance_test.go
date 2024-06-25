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

func TestBalanceGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewBalanceGenerator([]*config.Chain{})
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
					{Amount: 100000, Denom: "uatom"},
					{Amount: 200000, Denom: "ustake"},
				},
			},
			"consumer": {
				"validator": {
					{Amount: 100000, Denom: "untrn"},
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
			ConsumerChains: []*config.ConsumerChain{{
				Name: "consumer",
				Denoms: config.DenomInfos{
					{Denom: "untrn", DisplayDenom: "ntrn", DenomExponent: 6},
				},
			}},
		},
		{
			Name:       "chain2",
			Validators: []config.Validator{{Address: "validator"}, {Address: "validator2"}},
			ConsumerChains: []*config.ConsumerChain{{
				Name: "consumer2",
			}},
		},
	}
	generator := NewBalanceGenerator(chains)
	results := generator.Generate(state)
	assert.Len(t, results, 1)

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
	assert.InEpsilon(t, 0.1, testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain":   "consumer",
		"address": "validator",
		"denom":   "ntrn",
	})), 0.01)
	assert.InEpsilon(t, float64(200000), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain":   "consumer",
		"address": "validator",
		"denom":   "ustake",
	})), 0.01)
}
