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

func TestBalanceFetcherBase(t *testing.T) {
	t.Parallel()

	fetcher := NewBalanceFetcher(
		logger.GetDefaultLogger(),
		[]*config.Chain{},
		map[string]*tendermint.RPCWithConsumers{},
		tracing.InitNoopTracer(),
	)

	assert.NotNil(t, fetcher)
	assert.Equal(t, constants.FetcherNameBalance, fetcher.Name())
}

func TestBalanceFetcherNoBechWalletPrefix(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:        "chain",
		LCDEndpoint: "example",
		Validators:  []config.Validator{{Address: "validator"}},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetDefaultLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &BalanceFetcher{
		Logger: *logger.GetDefaultLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Empty(t, queries)

	balanceData, ok := data.(BalanceData)
	assert.True(t, ok)

	chainData, ok := balanceData.Balances["chain"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

func TestBalanceFetcherInvalidBechWalletPrefix(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "example",
		BechWalletPrefix: "test",
		Validators:       []config.Validator{{Address: "validator"}},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetDefaultLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &BalanceFetcher{
		Logger: *logger.GetDefaultLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Empty(t, queries)

	balanceData, ok := data.(BalanceData)
	assert.True(t, ok)

	chainData, ok := balanceData.Balances["chain"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

func TestBalanceFetcherQueryDisabled(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "example",
		BechWalletPrefix: "test",
		Validators:       []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
		Queries:          map[string]bool{"balance": false},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetDefaultLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &BalanceFetcher{
		Logger: *logger.GetDefaultLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Empty(t, queries)

	balanceData, ok := data.(BalanceData)
	assert.True(t, ok)

	chainData, ok := balanceData.Balances["chain"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestBalanceFetcherQueryError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/bank/v1beta1/balances/cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2",
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
			*logger.GetDefaultLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &BalanceFetcher{
		Logger: *logger.GetDefaultLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	balanceData, ok := data.(BalanceData)
	assert.True(t, ok)

	chainData, ok := balanceData.Balances["chain"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestBalanceFetcherNodeError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/bank/v1beta1/balances/cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2",
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
			*logger.GetDefaultLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &BalanceFetcher{
		Logger: *logger.GetDefaultLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	balanceData, ok := data.(BalanceData)
	assert.True(t, ok)

	chainData, ok := balanceData.Balances["chain"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestBalanceFetcherQuerySuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/bank/v1beta1/balances/cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("balances.json")),
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
			*logger.GetDefaultLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &BalanceFetcher{
		Logger: *logger.GetDefaultLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.True(t, queries[0].Success)

	balanceData, ok := data.(BalanceData)
	assert.True(t, ok)

	chainData, ok := balanceData.Balances["chain"]
	assert.True(t, ok)

	validatorData, ok := chainData["cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"]
	assert.True(t, ok)
	assert.Len(t, validatorData, 1)
	assert.InEpsilon(t, float64(596250), validatorData[0].Amount, 0.01)
	assert.Equal(t, "uatom", validatorData[0].Denom)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestBalanceFetcherConsumer(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.neutron.quokkastake.io/cosmos/bank/v1beta1/balances/neutron1xqz9pemz5e5zycaa89kys5aw6m8rhgsv07va3d",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("balances.json")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators:       []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
		Queries:          map[string]bool{"balance": false},
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
			*logger.GetDefaultLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &BalanceFetcher{
		Logger: *logger.GetDefaultLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.True(t, queries[0].Success)

	balanceData, ok := data.(BalanceData)
	assert.True(t, ok)

	chainData, ok := balanceData.Balances["consumer"]
	assert.True(t, ok)

	validatorData, ok := chainData["cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"]
	assert.True(t, ok)
	assert.Len(t, validatorData, 1)
	assert.InEpsilon(t, float64(596250), validatorData[0].Amount, 0.01)
	assert.Equal(t, "uatom", validatorData[0].Denom)
}
