package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
)

func TestNodeInfoGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.State{}
	generator := NewNodeInfoGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestNodeInfoGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.State{}
	state.Set(constants.FetcherNameNodeInfo, fetchers.NodeInfoData{
		NodeInfos: map[string]*types.NodeInfoResponse{
			"chain": {
				DefaultNodeInfo: types.DefaultNodeInfo{Network: "network", Version: "version"},
				ApplicationVersion: types.ApplicationVersion{
					Name:             "name",
					AppName:          "appname",
					Version:          "1.2.3",
					CosmosSDKVersion: "0.50.7",
				},
			},
		},
	})

	generator := NewNodeInfoGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(1), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain":              "chain",
		"chain_id":           "network",
		"cosmos_sdk_version": "0.50.7",
		"tendermint_version": "version",
		"app_version":        "1.2.3",
		"name":               "name",
		"app_name":           "appname",
	})), 0.01)
}
