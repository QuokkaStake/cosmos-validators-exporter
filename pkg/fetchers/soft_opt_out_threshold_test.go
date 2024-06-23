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

func TestSoftOptOutThresholdFetcherBase(t *testing.T) {
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
	fetcher := NewSoftOptOutThresholdFetcher(
		logger.GetNopLogger(),
		chains,
		rpcs,
		tracing.InitNoopTracer(),
	)

	assert.NotNil(t, fetcher)
	assert.Equal(t, constants.FetcherNameSoftOptOutThreshold, fetcher.Name())
}

func TestSoftOptOutThresholdFetcherQueryDisabled(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "example",
		BechWalletPrefix: "test",
		Validators:       []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:             "consumer",
				LCDEndpoint:      "https://api.neutron.quokkastake.io",
				BechWalletPrefix: "neutron",
				Queries:          map[string]bool{"params": false},
			},
		},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &SoftOptOutThresholdFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Empty(t, queries)

	paramsData, ok := data.(SoftOptOutThresholdData)
	assert.True(t, ok)
	assert.Empty(t, paramsData.Thresholds)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestSoftOptOutThresholdFetcherQueryInvalidJson(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.neutron.quokkastake.io/cosmos/params/v1beta1/params?subspace=ccvconsumer&key=SoftOptOutThreshold",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("soft-opt-out-threshold-invalid.json")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.neutron.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators:       []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:             "consumer",
				LCDEndpoint:      "https://api.neutron.quokkastake.io",
				BechWalletPrefix: "neutron",
			},
		},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &SoftOptOutThresholdFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	paramsData, ok := data.(SoftOptOutThresholdData)
	assert.True(t, ok)
	assert.Empty(t, paramsData.Thresholds)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestSoftOptOutThresholdFetcherQueryError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.neutron.quokkastake.io/cosmos/params/v1beta1/params?subspace=ccvconsumer&key=SoftOptOutThreshold",
		httpmock.NewErrorResponder(errors.New("error")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.neutron.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators:       []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:             "consumer",
				LCDEndpoint:      "https://api.neutron.quokkastake.io",
				BechWalletPrefix: "neutron",
			},
		},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &SoftOptOutThresholdFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	paramsData, ok := data.(SoftOptOutThresholdData)
	assert.True(t, ok)
	assert.Empty(t, paramsData.Thresholds)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestSoftOptOutThresholdFetcherNodeError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.neutron.quokkastake.io/cosmos/params/v1beta1/params?subspace=ccvconsumer&key=SoftOptOutThreshold",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("error.json")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.neutron.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators:       []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:             "consumer",
				LCDEndpoint:      "https://api.neutron.quokkastake.io",
				BechWalletPrefix: "neutron",
			},
		},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &SoftOptOutThresholdFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	paramsData, ok := data.(SoftOptOutThresholdData)
	assert.True(t, ok)
	assert.Empty(t, paramsData.Thresholds)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestSoftOptOutThresholdFetcherSuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.neutron.quokkastake.io/cosmos/params/v1beta1/params?subspace=ccvconsumer&key=SoftOptOutThreshold",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("soft-opt-out-threshold.json")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.neutron.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators:       []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
		Queries:          map[string]bool{"params": false},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:             "consumer",
				LCDEndpoint:      "https://api.neutron.quokkastake.io",
				BechWalletPrefix: "neutron",
			},
		},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &SoftOptOutThresholdFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.True(t, queries[0].Success)

	paramsData, ok := data.(SoftOptOutThresholdData)
	assert.True(t, ok)

	chainData, ok := paramsData.Thresholds["consumer"]
	assert.True(t, ok)
	assert.InEpsilon(t, 0.05, chainData, 0.01)
}
