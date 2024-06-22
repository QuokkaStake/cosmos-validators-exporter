package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSigningInfoGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewSigningInfoGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestSigningInfoGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameSigningInfo, fetchers.SigningInfoData{
		SigningInfos: map[string]map[string]*types.SigningInfoResponse{
			"chain": {
				"validator": &types.SigningInfoResponse{
					ValSigningInfo: types.SigningInfo{
						MissedBlocksCounter: "100",
					},
				},
			},
		},
	})

	generator := NewSigningInfoGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)
}
