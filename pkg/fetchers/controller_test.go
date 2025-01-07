package fetchers

import (
	"context"
	"main/assets"
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"main/pkg/tendermint"
	"main/pkg/tracing"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

//nolint:paralleltest // disabled due to httpmock usage
func TestControllerFetcherEnabled(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/base/tendermint/v1beta1/node_info",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("node-info.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/distribution/v1beta1/delegators/cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2/rewards/cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rewards.json")),
	)

	chains := []*configPkg.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators:       []configPkg.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*loggerPkg.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher1 := NewNodeInfoFetcher(
		loggerPkg.GetNopLogger(),
		chains,
		rpcs,
		tracing.InitNoopTracer(),
	)
	fetcher2 := NewRewardsFetcher(
		loggerPkg.GetNopLogger(),
		chains,
		rpcs,
		tracing.InitNoopTracer(),
	)
	logger := loggerPkg.GetNopLogger()
	controller := NewController(Fetchers{fetcher1, fetcher2}, logger)

	data, queryInfos := controller.Fetch(context.Background())
	assert.Len(t, queryInfos, 2)
	assert.Len(t, data, 2)
}
