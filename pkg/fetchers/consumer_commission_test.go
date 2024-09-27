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

func TestConsumerCommissionFetcherBase(t *testing.T) {
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
	fetcher := NewConsumerCommissionFetcher(
		logger.GetNopLogger(),
		chains,
		rpcs,
		tracing.InitNoopTracer(),
	)

	assert.NotNil(t, fetcher)
	assert.Equal(t, constants.FetcherNameConsumerCommission, fetcher.Name())
}

func TestConsumerCommissionFetcherNoConsensusAddress(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "example",
		BechWalletPrefix: "test",
		Validators: []config.Validator{{
			Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
		}},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:       "consumer",
				ConsumerID: "0",
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
	fetcher := &ConsumerCommissionFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Empty(t, queries)

	commissionsData, ok := data.(ConsumerCommissionData)
	assert.True(t, ok)

	chainData, ok := commissionsData.Commissions["consumer"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

func TestConsumerCommissionFetcherQueryDisabled(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "example",
		BechWalletPrefix: "test",
		Validators: []config.Validator{{
			Address:          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
			ConsensusAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		}},
		Queries: map[string]bool{"consumer-commission": false},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:       "consumer",
				ConsumerID: "0",
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
	fetcher := &ConsumerCommissionFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Empty(t, queries)

	commissionsData, ok := data.(ConsumerCommissionData)
	assert.True(t, ok)

	chainData, ok := commissionsData.Commissions["consumer"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestConsumerCommissionFetcherQueryError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/interchain_security/ccv/provider/consumer_commission_rate/0/cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		httpmock.NewErrorResponder(errors.New("error")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators: []config.Validator{{
			Address:          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
			ConsensusAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		}},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:       "consumer",
				ConsumerID: "0",
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
	fetcher := &ConsumerCommissionFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	commissionsData, ok := data.(ConsumerCommissionData)
	assert.True(t, ok)

	chainData, ok := commissionsData.Commissions["consumer"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestConsumerCommissionFetcherNodeError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/interchain_security/ccv/provider/consumer_commission_rate/0/cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("error.json")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators: []config.Validator{{
			Address:          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
			ConsensusAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		}},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:       "consumer",
				ConsumerID: "0",
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
	fetcher := &ConsumerCommissionFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	commissionsData, ok := data.(ConsumerCommissionData)
	assert.True(t, ok)

	chainData, ok := commissionsData.Commissions["consumer"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestConsumerCommissionFetcherQuerySuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/interchain_security/ccv/provider/consumer_commission_rate/0/cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("consumer-commission.json")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators: []config.Validator{{
			Address:          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
			ConsensusAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		}},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:       "consumer",
				ConsumerID: "0",
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
	fetcher := &ConsumerCommissionFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.True(t, queries[0].Success)

	commissionsData, ok := data.(ConsumerCommissionData)
	assert.True(t, ok)

	chainData, ok := commissionsData.Commissions["consumer"]
	assert.True(t, ok)
	assert.NotNil(t, chainData)

	validatorData, ok := chainData["cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"]
	assert.True(t, ok)
	assert.NotNil(t, validatorData)
	assert.InEpsilon(t, 0.1, validatorData.Rate.MustFloat64(), 0.01)
}
