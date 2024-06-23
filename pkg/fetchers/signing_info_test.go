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

func TestSigningInfoFetcherBase(t *testing.T) {
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
	fetcher := NewSigningInfoFetcher(
		logger.GetNopLogger(),
		chains,
		rpcs,
		tracing.InitNoopTracer(),
	)

	assert.NotNil(t, fetcher)
	assert.Equal(t, constants.FetcherNameSigningInfo, fetcher.Name())
}

func TestSigningInfoFetcherNoValcons(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "example",
		BechWalletPrefix: "test",
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
	fetcher := &SigningInfoFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Empty(t, queries)

	paramsData, ok := data.(SigningInfoData)
	assert.True(t, ok)

	chainData, ok := paramsData.SigningInfos["chain"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

func TestSigningInfoFetcherQueryDisabled(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "example",
		BechWalletPrefix: "test",
		Validators: []config.Validator{{
			Address:          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
			ConsensusAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		}},
		Queries: map[string]bool{"signing-info": false},
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &SigningInfoFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Empty(t, queries)

	paramsData, ok := data.(SigningInfoData)
	assert.True(t, ok)

	chainData, ok := paramsData.SigningInfos["chain"]
	assert.True(t, ok)

	validatorData, ok := chainData["cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"]
	assert.True(t, ok)
	assert.Nil(t, validatorData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestSigningInfoFetcherQueryError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/slashing/v1beta1/signing_infos/cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
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
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &SigningInfoFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	paramsData, ok := data.(SigningInfoData)
	assert.True(t, ok)

	chainData, ok := paramsData.SigningInfos["chain"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestSigningInfoFetcherNodeError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/slashing/v1beta1/signing_infos/cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
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
	}}
	rpcs := map[string]*tendermint.RPCWithConsumers{
		"chain": tendermint.RPCWithConsumersFromChain(
			chains[0],
			10,
			*logger.GetNopLogger(),
			tracing.InitNoopTracer(),
		),
	}
	fetcher := &SigningInfoFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	paramsData, ok := data.(SigningInfoData)
	assert.True(t, ok)

	chainData, ok := paramsData.SigningInfos["chain"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestSigningInfoFetcherQuerySuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/slashing/v1beta1/signing_infos/cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("signing-info.json")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
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
	fetcher := &SigningInfoFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.True(t, queries[0].Success)

	infosData, ok := data.(SigningInfoData)
	assert.True(t, ok)

	chainData, ok := infosData.SigningInfos["chain"]
	assert.True(t, ok)

	validatorData, ok := chainData["cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"]
	assert.True(t, ok)
	assert.Equal(t, int64(8), validatorData.ValSigningInfo.MissedBlocksCounter.Int64(), 0.01)
}

func TestSigningInfoFetcherConsumerNoBechConsensusPrefix(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators: []config.Validator{{
			Address:          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
			ConsensusAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		}},
		Queries: map[string]bool{"signing-info": false},
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
	fetcher := &SigningInfoFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Empty(t, queries)

	paramsData, ok := data.(SigningInfoData)
	assert.True(t, ok)

	chainData, ok := paramsData.SigningInfos["chain"]
	assert.True(t, ok)

	validatorData, ok := chainData["cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"]
	assert.True(t, ok)
	assert.Nil(t, validatorData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestSigningInfoFetcherConsumerAssignedKeyQueryError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/interchain_security/ccv/provider/validator_consumer_addr?chain_id=consumer&provider_address=cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
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
		Queries: map[string]bool{"signing-info": false},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:                "consumer",
				ChainID:             "consumer",
				LCDEndpoint:         "https://api.neutron.quokkastake.io",
				BechWalletPrefix:    "neutron",
				BechConsensusPrefix: "neutronvalcons",
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
	fetcher := &SigningInfoFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	infosData, ok := data.(SigningInfoData)
	assert.True(t, ok)

	chainData, ok := infosData.SigningInfos["chain"]
	assert.True(t, ok)

	validatorData, ok := chainData["cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"]
	assert.True(t, ok)
	assert.Nil(t, validatorData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestSigningInfoFetcherConsumerAssignedKeyNodeError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/interchain_security/ccv/provider/validator_consumer_addr?chain_id=consumer&provider_address=cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
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
		Queries: map[string]bool{"signing-info": false},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:                "consumer",
				ChainID:             "consumer",
				LCDEndpoint:         "https://api.neutron.quokkastake.io",
				BechWalletPrefix:    "neutron",
				BechConsensusPrefix: "neutronvalcons",
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
	fetcher := &SigningInfoFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	infosData, ok := data.(SigningInfoData)
	assert.True(t, ok)

	chainData, ok := infosData.SigningInfos["chain"]
	assert.True(t, ok)

	validatorData, ok := chainData["cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"]
	assert.True(t, ok)
	assert.Nil(t, validatorData)
}

func TestSigningInfoFetcherConsumerAssignedKeyInvalidValcons(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators: []config.Validator{{
			Address:          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
			ConsensusAddress: "test",
		}},
		Queries: map[string]bool{"signing-info": false, "assigned-key": false},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:                "consumer",
				LCDEndpoint:         "https://api.neutron.quokkastake.io",
				BechWalletPrefix:    "neutron",
				BechConsensusPrefix: "neutronvalcons",
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
	fetcher := &SigningInfoFetcher{
		Logger: *logger.GetDefaultLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Empty(t, queries)

	infosData, ok := data.(SigningInfoData)
	assert.True(t, ok)

	chainData, ok := infosData.SigningInfos["chain"]
	assert.True(t, ok)

	validatorData, ok := chainData["cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"]
	assert.True(t, ok)
	assert.Nil(t, validatorData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestSigningInfoFetcherConsumerInvalidValoper(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/interchain_security/ccv/provider/validator_consumer_addr?chain_id=consumer&provider_address=cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("assigned-key.json")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators: []config.Validator{{
			Address:          "validator",
			ConsensusAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		}},
		Queries: map[string]bool{"signing-info": false},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:                "consumer",
				ChainID:             "consumer",
				LCDEndpoint:         "https://api.neutron.quokkastake.io",
				BechWalletPrefix:    "neutron",
				BechConsensusPrefix: "neutronvalcons",
				BechValidatorPrefix: "neutronvaloper",
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
	fetcher := &SigningInfoFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.True(t, queries[0].Success)

	infosData, ok := data.(SigningInfoData)
	assert.True(t, ok)

	chainData, ok := infosData.SigningInfos["consumer"]
	assert.True(t, ok)
	assert.Empty(t, chainData)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestSigningInfoFetcherConsumerSuccessWithAssignedKey(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/interchain_security/ccv/provider/validator_consumer_addr?chain_id=consumer&provider_address=cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("assigned-key.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.neutron.quokkastake.io/cosmos/slashing/v1beta1/signing_infos/neutronvalcons1w426hkttrwrve9mj77ld67lzgx5u9m8plhmwc6",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("signing-info.json")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators: []config.Validator{{
			Address:          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
			ConsensusAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		}},
		Queries: map[string]bool{"signing-info": false},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:                "consumer",
				ChainID:             "consumer",
				LCDEndpoint:         "https://api.neutron.quokkastake.io",
				BechWalletPrefix:    "neutron",
				BechConsensusPrefix: "neutronvalcons",
				BechValidatorPrefix: "neutronvaloper",
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
	fetcher := &SigningInfoFetcher{
		Logger: *logger.GetNopLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 2)
	assert.True(t, queries[0].Success)
	assert.True(t, queries[0].Success)

	infosData, ok := data.(SigningInfoData)
	assert.True(t, ok)

	chainData, ok := infosData.SigningInfos["consumer"]
	assert.True(t, ok)

	validatorData, ok := chainData["neutronvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsv4r4yhf"]
	assert.True(t, ok)
	assert.Equal(t, int64(8), validatorData.ValSigningInfo.MissedBlocksCounter.Int64(), 0.01)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestSigningInfoFetcherConsumerSuccessWithoutAssignedKey(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/interchain_security/ccv/provider/validator_consumer_addr?chain_id=consumer&provider_address=cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("assigned-key-empty.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.neutron.quokkastake.io/cosmos/slashing/v1beta1/signing_infos/neutronvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnc0mxjg",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("signing-info.json")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators: []config.Validator{{
			Address:          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
			ConsensusAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		}},
		Queries: map[string]bool{"signing-info": false},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:                "consumer",
				ChainID:             "consumer",
				LCDEndpoint:         "https://api.neutron.quokkastake.io",
				BechWalletPrefix:    "neutron",
				BechConsensusPrefix: "neutronvalcons",
				BechValidatorPrefix: "neutronvaloper",
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
	fetcher := &SigningInfoFetcher{
		Logger: *logger.GetDefaultLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 2)
	assert.True(t, queries[0].Success)
	assert.True(t, queries[0].Success)

	infosData, ok := data.(SigningInfoData)
	assert.True(t, ok)

	chainData, ok := infosData.SigningInfos["consumer"]
	assert.True(t, ok)

	validatorData, ok := chainData["neutronvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsv4r4yhf"]
	assert.True(t, ok)
	assert.Equal(t, int64(8), validatorData.ValSigningInfo.MissedBlocksCounter.Int64(), 0.01)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestSigningInfoFetcherConsumerSuccessWithAssignedKeyQueryDisabled(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.neutron.quokkastake.io/cosmos/slashing/v1beta1/signing_infos/neutronvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnc0mxjg",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("signing-info.json")),
	)

	chains := []*config.Chain{{
		Name:             "chain",
		LCDEndpoint:      "https://api.cosmos.quokkastake.io",
		BechWalletPrefix: "cosmos",
		Validators: []config.Validator{{
			Address:          "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
			ConsensusAddress: "cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		}},
		Queries: map[string]bool{"signing-info": false, "assigned-key": false},
		ConsumerChains: []*config.ConsumerChain{
			{
				Name:                "consumer",
				ChainID:             "consumer",
				LCDEndpoint:         "https://api.neutron.quokkastake.io",
				BechWalletPrefix:    "neutron",
				BechConsensusPrefix: "neutronvalcons",
				BechValidatorPrefix: "neutronvaloper",
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
	fetcher := &SigningInfoFetcher{
		Logger: *logger.GetDefaultLogger(),
		Chains: chains,
		RPCs:   rpcs,
		Tracer: tracing.InitNoopTracer(),
	}
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.True(t, queries[0].Success)

	infosData, ok := data.(SigningInfoData)
	assert.True(t, ok)

	chainData, ok := infosData.SigningInfos["consumer"]
	assert.True(t, ok)

	validatorData, ok := chainData["neutronvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsv4r4yhf"]
	assert.True(t, ok)
	assert.Equal(t, int64(8), validatorData.ValSigningInfo.MissedBlocksCounter.Int64(), 0.01)
}
