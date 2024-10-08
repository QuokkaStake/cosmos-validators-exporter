package fetchers

import (
	"context"
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestPriceFetcherBase(t *testing.T) {
	t.Parallel()

	chains := []*configPkg.Chain{
		{Name: "chain1", LCDEndpoint: "example1"},
		{Name: "chain2", LCDEndpoint: "example2"},
	}
	config := &configPkg.Config{Chains: chains}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()

	fetcher := NewPriceFetcher(
		logger,
		config,
		tracer,
	)

	assert.NotNil(t, fetcher)
	assert.Equal(t, constants.FetcherNamePrice, fetcher.Name())
}

//nolint:paralleltest // disabled due to httpmock usage
func TestPriceFetcherProviderCoingeckoError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.coingecko.com/api/v3/simple/price?ids=cosmos&vs_currencies=usd",
		httpmock.NewErrorResponder(errors.New("error")),
	)

	chains := []*configPkg.Chain{{
		Name:   "chain",
		Denoms: configPkg.DenomInfos{{Denom: "uatom", DisplayDenom: "atom", CoingeckoCurrency: "cosmos"}},
	}}
	config := &configPkg.Config{Chains: chains}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()

	fetcher := NewPriceFetcher(
		logger,
		config,
		tracer,
	)
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	balanceData, ok := data.(PriceData)
	assert.True(t, ok)
	assert.Empty(t, balanceData.Prices)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestPriceFetcherProviderCoingeckoSuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.coingecko.com/api/v3/simple/price?ids=cosmos&vs_currencies=usd",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("coingecko.json")),
	)

	chains := []*configPkg.Chain{{
		Name:   "chain",
		Denoms: configPkg.DenomInfos{{Denom: "uatom", DisplayDenom: "atom", CoingeckoCurrency: "cosmos"}},
	}}
	config := &configPkg.Config{Chains: chains}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()

	fetcher := NewPriceFetcher(
		logger,
		config,
		tracer,
	)
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.True(t, queries[0].Success)

	balanceData, ok := data.(PriceData)
	assert.True(t, ok)

	chainData, ok := balanceData.Prices["chain"]
	assert.True(t, ok)

	denomData, ok := chainData["atom"]
	assert.True(t, ok)
	assert.InEpsilon(t, 6.71, denomData.Value, 0.01)
	assert.Equal(t, constants.CoingeckoBaseCurrency, denomData.BaseCurrency)
	assert.Equal(t, constants.PriceFetcherNameCoingecko, denomData.Source)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestPriceFetcherConsumerCoingeckoSuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.coingecko.com/api/v3/simple/price?ids=cosmos,test&vs_currencies=usd",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("coingecko.json")),
	)

	chains := []*configPkg.Chain{{
		Name: "chain",
		ConsumerChains: []*configPkg.ConsumerChain{{
			Name: "consumer",
			Denoms: configPkg.DenomInfos{
				{Denom: "uatom", DisplayDenom: "atom", CoingeckoCurrency: "cosmos"},
				{Denom: "utest", DisplayDenom: "test", CoingeckoCurrency: "test"},
				{Denom: "ustake", DisplayDenom: "stake"},
			},
		}},
	}}
	config := &configPkg.Config{Chains: chains}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()

	fetcher := NewPriceFetcher(
		logger,
		config,
		tracer,
	)
	data, queries := fetcher.Fetch(context.Background())
	assert.Len(t, queries, 1)
	assert.True(t, queries[0].Success)

	balanceData, ok := data.(PriceData)
	assert.True(t, ok)

	chainData, ok := balanceData.Prices["consumer"]
	assert.True(t, ok)

	denomData, ok := chainData["atom"]
	assert.True(t, ok)
	assert.InEpsilon(t, 6.71, denomData.Value, 0.01)
	assert.Equal(t, constants.CoingeckoBaseCurrency, denomData.BaseCurrency)
	assert.Equal(t, constants.PriceFetcherNameCoingecko, denomData.Source)
}
