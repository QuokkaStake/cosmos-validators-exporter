package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"cosmossdk.io/math"

	"github.com/stretchr/testify/assert"
)

func TestSlashingParamsGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.State{}
	generator := NewSlashingParamsGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestSlashingParamsGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.State{}
	state.Set(constants.FetcherNameSlashingParams, fetchers.SlashingParamsData{
		Params: map[string]*types.SlashingParamsResponse{
			"chain": {
				SlashingParams: types.SlashingParams{
					SignedBlocksWindow: math.NewInt(100),
				},
			},
		},
	})

	generator := NewSlashingParamsGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(100), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)
}
