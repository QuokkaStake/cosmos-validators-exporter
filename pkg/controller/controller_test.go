package controller

import (
	"context"
	fetchersPkg "main/pkg/fetchers"
	loggerPkg "main/pkg/logger"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestControllerFetcherEnabled(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	controller := NewController(fetchersPkg.Fetchers{
		&fetchersPkg.StubFetcher1{},
		&fetchersPkg.StubFetcher2{},
	}, logger)

	data, queryInfos := controller.Fetch(context.Background())
	assert.Empty(t, queryInfos)
	assert.Len(t, data, 2)
}
