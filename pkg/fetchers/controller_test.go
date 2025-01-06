package fetchers

import (
	"context"
	"main/assets"
	"main/pkg/clients/tendermint"
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestControllerFetcherNotEnabled(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetDefaultLogger()
	tracer := tracing.InitNoopTracer()
	fetcher := NewNodeStatusFetcher(*logger, nil, tracer)

	controller := NewController(Fetchers{fetcher}, *logger, "chain")

	data, queryInfos := controller.Fetch(context.Background())
	assert.Empty(t, queryInfos)
	assert.Empty(t, data)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestControllerFetcherEnabled(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/status",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("status.json")),
	)

	config := configPkg.TendermintConfig{
		Address: "https://example.com",
	}

	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	fetcher := NewNodeStatusFetcher(*logger, client, tracer)
	controller := NewController(Fetchers{fetcher}, *logger, "chain")

	data, queryInfos := controller.Fetch(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.Len(t, data, 1)
}
