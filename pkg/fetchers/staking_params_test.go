package fetchers

import (
	"context"
	"errors"
	"main/assets"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/logger"
	"main/pkg/tendermint"
	"main/pkg/tracing"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jarcoal/httpmock"
)

func TestStakingParamsFetcherBase(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{
		{Name: "chain1", LCDEndpoint: "example1"},
		{Name: "chain2", LCDEndpoint: "example2"},
	}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain1": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
		"chain2": tendermint.RPCWithConsumersFromChain(
			chains[1],
			10,
			*logger.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := NewStakingParamsFetcher(
		logger.GetNopLogger(),
		chains,
		rpcs,
		tracing.InitNoopTracer(),
	)

	assert.NotNil(t, fetcher)
	assert.Equal(t, constants.FetcherNameStakingParams, fetcher.Name())
}

func TestStakingParamsFetcherQueryDisabled(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "example",
		BechWalletPrefix: "test",
		Validators:       []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
		Queries:          map[string]bool{"staking-params": false},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetDefaultLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &StakingParamsFetcher{
		Logger: *logger.GetDefaultLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Empty(t, queries)

	paramsData, ok := data.(StakingParamsData)
	assert.True(t, ok)
	assert.Empty(t, paramsData.Params)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestStakingParamsFetcherQueryError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/staking/v1beta1/params",
		httpmock.NewErrorResponder(errors.New("error")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators:       []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &StakingParamsFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	paramsData, ok := data.(StakingParamsData)
	assert.True(t, ok)
	assert.Empty(t, paramsData.Params)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestStakingParamsFetcherNodeError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/staking/v1beta1/params",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("error.json")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators:       []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &StakingParamsFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	paramsData, ok := data.(StakingParamsData)
	assert.True(t, ok)
	assert.Empty(t, paramsData.Params)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestStakingParamsFetcherQuerySuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/staking/v1beta1/params",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("staking-params.json")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators:       []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
		ConsumerChains: []*config.ConsumerChain{{
			Name: "consumer",
		}},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &StakingParamsFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.True(t, queries[0].Success)

	paramsData, ok := data.(StakingParamsData)
	assert.True(t, ok)

	chainData, ok := paramsData.Params["chain"]
	assert.True(t, ok)
	assert.NotNil(t, chainData)
	assert.Equal(t, 180, chainData.StakingParams.MaxValidators)

	consumerData, ok := paramsData.Params["consumer"]
	assert.True(t, ok)
	assert.NotNil(t, consumerData)
	assert.Equal(t, 180, consumerData.StakingParams.MaxValidators)
}
