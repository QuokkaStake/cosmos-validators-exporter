package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriceGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewPriceGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestPriceGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNamePrice, fetchers.PriceData{
		Prices: map[string]map[string]float64{
			"chain": {
				"denom": 0.01,
			},
		},
	})

	generator := NewPriceGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)
}
