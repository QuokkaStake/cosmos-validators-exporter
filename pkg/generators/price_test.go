package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
)

func TestPriceGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.State{}
	generator := NewPriceGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestPriceGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.State{}
	state.Set(constants.FetcherNamePrice, fetchers.PriceData{
		Prices: map[string]map[string]fetchers.PriceInfo{
			"chain": {
				"denom": fetchers.PriceInfo{
					Value:        0.01,
					Source:       constants.PriceFetcherNameCoingecko,
					BaseCurrency: constants.CoingeckoBaseCurrency,
				},
			},
		},
	})

	generator := NewPriceGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, 0.01, testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain":         "chain",
		"denom":         "denom",
		"source":        string(constants.PriceFetcherNameCoingecko),
		"base_currency": constants.CoingeckoBaseCurrency,
	})), 0.01)
}
