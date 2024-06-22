package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelfDelegationGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewSelfDelegationGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestSelfDelegationGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameSelfDelegation, fetchers.SelfDelegationData{
		Delegations: map[string]map[string]*types.Amount{
			"chain": {
				"validator": {Amount: 1, Denom: "denom"},
			},
		},
	})

	generator := NewSelfDelegationGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)
}