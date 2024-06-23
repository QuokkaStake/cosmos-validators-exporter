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

func TestCommissionFetcherBase(t *testing.T) {
	t.Parallel()

	fetcher := NewCommissionFetcher(
		logger.GetNopLogger(),
		[]*config.Chain{},
		map[string]*tendermint.RPCWithConsumers{},
		tracing.InitNoopTracer(),
	)

	assert.NotNil(t, fetcher)
	assert.Equal(t, constants.FetcherNameCommission, fetcher.Name())
}

func TestCommissionFetcherQueryDisabled(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "example",
		BechWalletPrefix: "test",
		Validators:       []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
		Queries:          map[string]bool{"commission": false},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &CommissionFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Empty(t, queries)

	balanceData, ok := data.(CommissionData)
	assert.True(t, ok)

	chainData, ok := balanceData.Commissions["chain"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestCommissionFetcherQueryError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/distribution/v1beta1/validators/cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e/commission",
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
	fetcher := &CommissionFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	balanceData, ok := data.(CommissionData)
	assert.True(t, ok)

	chainData, ok := balanceData.Commissions["chain"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestCommissionFetcherNodeError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/distribution/v1beta1/validators/cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e/commission",
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
	fetcher := &CommissionFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	balanceData, ok := data.(CommissionData)
	assert.True(t, ok)

	chainData, ok := balanceData.Commissions["chain"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestCommissionFetcherQuerySuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/distribution/v1beta1/validators/cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e/commission",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("commission.json")),
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
	fetcher := &CommissionFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.True(t, queries[0].Success)

	balanceData, ok := data.(CommissionData)
	assert.True(t, ok)

	chainData, ok := balanceData.Commissions["chain"]
	assert.True(t, ok)

	validatorData, ok := chainData["cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"]
	assert.True(t, ok)
	assert.Len(t, validatorData, 1)
	assert.InEpsilon(t, float64(35387184.119122794099577747), validatorData[0].Amount, 0.01)
	assert.Equal(t, "uatom", validatorData[0].Denom)
}
