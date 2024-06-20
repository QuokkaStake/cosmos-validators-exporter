package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBalanceGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewBalanceGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestBlockSetValidators(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameBalance, fetchers.BalanceData{
		Balances: map[string]map[string][]types.Amount{
			"chain": {
				"validator": {
					{Amount: 100, Denom: "uatom"},
					{Amount: 200, Denom: "ustake"},
				},
			},
		},
	})

	generator := NewBalanceGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)
}
