package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommissionGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewCommissionGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestCommissionGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameCommission, fetchers.CommissionData{
		Commissions: map[string]map[string][]types.Amount{
			"chain": {
				"validator": []types.Amount{
					{Amount: 1, Denom: "denom"},
				},
			},
		},
	})

	generator := NewCommissionGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)
}
