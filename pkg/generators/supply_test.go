package generators

import (
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"github.com/guregu/null/v5"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
)

func TestSupplyGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewSupplyGenerator([]*config.Chain{})
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestSupplyGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameSupply, fetchers.SupplyData{
		Supplies: map[string][]types.Amount{
			"chain": {
				{Amount: 100000, Denom: "uatom"},
				{Amount: 200000, Denom: "ustake"},
				{Amount: 300000, Denom: "uignored"},
			},
			"consumer": {
				{Amount: 100000, Denom: "untrn"},
				{Amount: 200000, Denom: "ustake"},
				{Amount: 300000, Denom: "uignored"},
			},
		},
	})

	chains := []*config.Chain{
		{
			Name:       "chain",
			Validators: []config.Validator{{Address: "validator"}, {Address: "validator2"}},
			Denoms: config.DenomInfos{
				{Denom: "uatom", DisplayDenom: "atom", DenomExponent: 6},
				{Denom: "uignored", Ignore: null.BoolFrom(true)},
			},
			ConsumerChains: []*config.ConsumerChain{{
				Name: "consumer",
				Denoms: config.DenomInfos{
					{Denom: "untrn", DisplayDenom: "ntrn", DenomExponent: 6},
					{Denom: "uignored", Ignore: null.BoolFrom(true)},
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
	generator := NewSupplyGenerator(chains)
	results := generator.Generate(state)
	assert.Len(t, results, 1)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 4, testutil.CollectAndCount(gauge))
	assert.InEpsilon(t, 0.1, testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain": "chain",
		"denom": "atom",
	})), 0.01)
	assert.InEpsilon(t, float64(200000), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain": "chain",
		"denom": "ustake",
	})), 0.01)
	assert.InEpsilon(t, 0.1, testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain": "consumer",
		"denom": "ntrn",
	})), 0.01)
	assert.InEpsilon(t, float64(200000), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain": "consumer",
		"denom": "ustake",
	})), 0.01)
}
