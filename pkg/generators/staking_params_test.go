package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStakingParamsGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewStakingParamsGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestStakingParamsGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
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
}
