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

func TestStakingParamsGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.State{}
	generator := NewStakingParamsGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestStakingParamsGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.State{}
	state.Set(constants.FetcherNameStakingParams, fetchers.StakingParamsData{
		Params: map[string]*types.StakingParamsResponse{
			"chain": {
				StakingParams: types.StakingParams{
					MaxValidators: 100,
				},
			},
		},
	})

	generator := NewStakingParamsGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(100), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)
}
