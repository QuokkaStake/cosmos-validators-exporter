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

	"github.com/guregu/null/v5"

	"github.com/stretchr/testify/assert"

	"github.com/jarcoal/httpmock"
)

func TestHasToValidateFetcherBase(t *testing.T) {
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
	fetcher := NewValidatorConsumersFetcher(
		logger.GetNopLogger(),
		chains,
		rpcs,
		tracing.InitNoopTracer(),
	)

	assert.NotNil(t, fetcher)
	assert.Equal(t, constants.FetcherNameValidatorConsumers, fetcher.Name())
}

func TestHasToValidateFetcherNotProvider(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "example",
		BechWalletPrefix: "test",
		IsProvider:       null.BoolFrom(false),
		Validators: []config.Validator{{
			Address:          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
			ConsensusAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
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
	fetcher := &ValidatorConsumersFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Empty(t, queries)

	validatorsData, ok := data.(ValidatorConsumersData)
	assert.True(t, ok)
	assert.Empty(t, validatorsData.Infos)
}

func TestHasToValidateFetcherNoConsensusKey(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "example",
		BechWalletPrefix: "test",
		IsProvider:       null.BoolFrom(true),
		Validators: []config.Validator{{
			Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
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
	fetcher := &ValidatorConsumersFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Empty(t, queries)

	validatorsData, ok := data.(ValidatorConsumersData)
	assert.True(t, ok)

	chainData, ok := validatorsData.Infos["chain"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

func TestHasToValidateFetcherQueryDisabled(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "example",
		BechWalletPrefix: "test",
		IsProvider:       null.BoolFrom(true),
		Validators: []config.Validator{{
			Address:          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
			ConsensusAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		}},
		Queries: map[string]bool{"validator-consumer-chains": false},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &ValidatorConsumersFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Empty(t, queries)

	validatorsData, ok := data.(ValidatorConsumersData)
	assert.True(t, ok)

	chainData, ok := validatorsData.Infos["chain"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestHasToValidateFetcherQueryError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/interchain_security/ccv/provider/consumer_chains_per_validator/cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		httpmock.NewErrorResponder(errors.New("error")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		IsProvider:       null.BoolFrom(true),
		Validators: []config.Validator{{
			Address:          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
			ConsensusAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
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
	fetcher := &ValidatorConsumersFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	validatorsData, ok := data.(ValidatorConsumersData)
	assert.True(t, ok)

	chainData, ok := validatorsData.Infos["chain"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestHasToValidateFetcherNodeError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/interchain_security/ccv/provider/consumer_chains_per_validator/cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("error.json")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		IsProvider:       null.BoolFrom(true),
		Validators: []config.Validator{{
			Address:          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
			ConsensusAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
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
	fetcher := &ValidatorConsumersFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	validatorsData, ok := data.(ValidatorConsumersData)
	assert.True(t, ok)

	chainData, ok := validatorsData.Infos["chain"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestHasToValidateFetcherQuerySuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/interchain_security/ccv/provider/consumer_chains_per_validator/cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("validator-consumers.json")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		IsProvider:       null.BoolFrom(true),
		Validators: []config.Validator{{
			Address:          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
			ConsensusAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
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
	fetcher := &ValidatorConsumersFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.True(t, queries[0].Success)

	validatorsData, ok := data.(ValidatorConsumersData)
	assert.True(t, ok)

	chainData, ok := validatorsData.Infos["chain"]
	assert.True(t, ok)

	validatorData, ok := chainData["cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"]
	assert.True(t, ok)
	assert.Len(t, validatorData, 2)
}
