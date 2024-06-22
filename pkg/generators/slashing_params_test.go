package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlashingParamsGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewSlashingParamsGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestSlashingParamsGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameSlashingParams, fetchers.SlashingParamsData{
		Params: map[string]*types.SlashingParamsResponse{
			"chain": {
				SlashingParams: types.SlashingParams{
					SignedBlocksWindow: "100",
				},
			},
		},
	})

	generator := NewSlashingParamsGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)
}
