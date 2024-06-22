package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeInfoGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewNodeInfoGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestNodeInfoGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameNodeInfo, fetchers.NodeInfoData{
		NodeInfos: map[string]*types.NodeInfoResponse{
			"chain": {
				DefaultNodeInfo: types.DefaultNodeInfo{Network: "network", Version: "version"},
				ApplicationVersion: types.ApplicationVersion{
					Name:             "app",
					AppName:          "app",
					Version:          "1.2.3",
					CosmosSDKVersion: "0.50.7",
				},
			},
		},
	})

	generator := NewNodeInfoGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)
}
