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

func TestValidatorsFetcherBase(t *testing.T) {
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
	fetcher := NewValidatorsFetcher(
		logger.GetNopLogger(),
		chains,
		rpcs,
		tracing.InitNoopTracer(),
	)

	assert.NotNil(t, fetcher)
	assert.Equal(t, constants.FetcherNameValidators, fetcher.Name())
}

func TestValidatorsFetcherQueryDisabled(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "example",
		BechWalletPrefix: "test",
		Validators:       []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
		Queries:          map[string]bool{"validators": false},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &ValidatorsFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Empty(t, queries)

	validatorsData, ok := data.(ValidatorsData)
	assert.True(t, ok)

	chainData, ok := validatorsData.Validators["chain"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestValidatorsFetcherQueryError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/staking/v1beta1/validators?pagination.count_total=true&pagination.limit=1000",
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
	fetcher := &ValidatorsFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	validatorsData, ok := data.(ValidatorsData)
	assert.True(t, ok)
	assert.Empty(t, validatorsData.Validators)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestValidatorsFetcherNodeError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/staking/v1beta1/validators?pagination.count_total=true&pagination.limit=1000",
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
	fetcher := &ValidatorsFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	validatorsData, ok := data.(ValidatorsData)
	assert.True(t, ok)
	assert.Empty(t, validatorsData.Validators)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestValidatorsFetcherQuerySuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/staking/v1beta1/validators?pagination.count_total=true&pagination.limit=1000",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("validators.json")),
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
	fetcher := &ValidatorsFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.True(t, queries[0].Success)

	validatorsData, ok := data.(ValidatorsData)
	assert.True(t, ok)

	chainData, ok := validatorsData.Validators["chain"]
	assert.True(t, ok)
	assert.Len(t, chainData.Validators, 542)
}
