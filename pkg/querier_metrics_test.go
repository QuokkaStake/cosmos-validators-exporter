package pkg

import (
	"context"
	"main/pkg/config"
	"main/pkg/types"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestQuerierMetrics(t *testing.T) {
	t.Parallel()

	queryInfos := []*types.QueryInfo{
		{Success: true, Chain: "chain", Duration: 2 * time.Second, URL: "url1"},
		{Success: true, Chain: "chain", Duration: 4 * time.Second, URL: "url2"},
		{Success: false, Chain: "chain", Duration: 6 * time.Second, URL: "url3"},
	}

	chains := []*config.Chain{
		{Name: "chain"},
		{Name: "chain2"},
	}

	generator := NewQueriesMetrics(chains, queryInfos)
	metrics := generator.GetMetrics(context.Background())
	assert.Len(t, metrics, 4)

	queriesCountGauge, ok := metrics[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 2, testutil.CollectAndCount(queriesCountGauge))
	assert.InDelta(t, 3, testutil.ToFloat64(queriesCountGauge.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)
	assert.Zero(t, testutil.ToFloat64(queriesCountGauge.With(prometheus.Labels{
		"chain": "chain2",
	})))

	queriesSuccess, ok := metrics[1].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 2, testutil.CollectAndCount(queriesSuccess))
	assert.InDelta(t, 2, testutil.ToFloat64(queriesSuccess.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)
	assert.Zero(t, testutil.ToFloat64(queriesSuccess.With(prometheus.Labels{
		"chain": "chain2",
	})))

	queriesFailed, ok := metrics[2].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 2, testutil.CollectAndCount(queriesFailed))
	assert.InDelta(t, 1, testutil.ToFloat64(queriesFailed.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)
	assert.Zero(t, testutil.ToFloat64(queriesFailed.With(prometheus.Labels{
		"chain": "chain2",
	})))

	timings, ok := metrics[3].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 3, testutil.CollectAndCount(timings))
	assert.InDelta(t, 2, testutil.ToFloat64(timings.With(prometheus.Labels{
		"chain": "chain",
		"url":   "url1",
	})), 0.01)
	assert.InDelta(t, 4, testutil.ToFloat64(timings.With(prometheus.Labels{
		"chain": "chain",
		"url":   "url2",
	})), 0.01)
	assert.InDelta(t, 6, testutil.ToFloat64(timings.With(prometheus.Labels{
		"chain": "chain",
		"url":   "url3",
	})), 0.01)
}
