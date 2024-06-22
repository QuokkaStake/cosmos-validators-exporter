package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRewardsGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewRewardsGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestRewardsGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameRewards, fetchers.RewardsData{
		Rewards: map[string]map[string][]types.Amount{
			"chain": {
				"validator": []types.Amount{
					{Amount: 1, Denom: "denom"},
				},
			},
		},
	})

	generator := NewRewardsGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)
}
