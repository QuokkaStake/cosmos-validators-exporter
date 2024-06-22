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

func TestSoftOptOutThresholdGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewSoftOptOutThresholdGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestSoftOptOutThresholdGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameSoftOptOutThreshold, fetchers.SoftOptOutThresholdData{
		Thresholds: map[string]float64{
			"chain": 0.5,
		},
	})

	generator := NewSoftOptOutThresholdGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, 0.5, testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)
}
