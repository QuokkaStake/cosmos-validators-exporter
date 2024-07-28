package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"testing"

	"cosmossdk.io/math"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
)

func TestInflationGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewInflationGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestInflationGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameInflation, fetchers.InflationData{
		Inflation: map[string]math.LegacyDec{
			"chain": math.LegacyMustNewDecFromStr("0.1"),
		},
	})

	generator := NewInflationGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, 0.1, testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)
}
