package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelegationsGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewDelegationsGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestDelegationsGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameDelegations, fetchers.DelegationsData{
		Delegations: map[string]map[string]uint64{
			"chain": {
				"validator": 100,
			},
		},
	})

	generator := NewDelegationsGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)
}
