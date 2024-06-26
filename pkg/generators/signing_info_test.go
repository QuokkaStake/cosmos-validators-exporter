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

func TestSigningInfoGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewSigningInfoGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestSigningInfoGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameSigningInfo, fetchers.SigningInfoData{
		SigningInfos: map[string]map[string]*types.SigningInfoResponse{
			"chain": {
				"validator": &types.SigningInfoResponse{
					ValSigningInfo: types.SigningInfo{
						MissedBlocksCounter: math.NewInt(100),
					},
				},
			},
		},
	})

	generator := NewSigningInfoGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(100), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain":   "chain",
		"address": "validator",
	})), 0.01)
}
