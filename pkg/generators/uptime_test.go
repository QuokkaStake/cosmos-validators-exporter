package generators

import (
	statePkg "main/pkg/state"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
)

func TestUptimeGenerator(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewUptimeGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.NotEmpty(t, testutil.ToFloat64(gauge))
}
