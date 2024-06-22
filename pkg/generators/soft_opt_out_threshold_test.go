package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"testing"

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
}
