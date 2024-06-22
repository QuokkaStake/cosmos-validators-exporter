package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnbondsGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewUnbondsGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestUnbondsGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameUnbonds, fetchers.UnbondsData{
		Unbonds: map[string]map[string]uint64{
			"chain": {
				"validator": 100,
			},
		},
	})

	generator := NewUnbondsGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)
}
