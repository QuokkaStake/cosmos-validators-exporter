package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"cosmossdk.io/math"

	"github.com/stretchr/testify/assert"
)

func TestConsumerInfoGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewConsumerInfoGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestConsumerInfoGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameConsumerInfo, fetchers.ConsumerInfoData{
		Info: map[string]*types.ConsumerInfoResponse{
			"chain": {
				Chains: []types.ConsumerChainInfo{
					{ChainID: "chain-id", TopN: 1, MinPowerInTopN: math.NewInt(100)},
				},
			},
		},
	})

	generator := NewConsumerInfoGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)
}
