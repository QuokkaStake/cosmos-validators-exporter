package types

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type QueryInfo struct {
	Chain    string
	URL      string
	Duration time.Duration
	Success  bool
}

func (q *ValidatorQuery) GetSuccessfulQueriesCount() float64 {
	var count int64 = 0

	for _, query := range q.Queries {
		if query.Success {
			count++
		}
	}

	return float64(count)
}

func (q *ValidatorQuery) GetFailedQueriesCount() float64 {
	return float64(len(q.Queries)) - q.GetSuccessfulQueriesCount()
}

type Amount struct {
	Amount float64
	Denom  string
}

type Querier interface {
	GetMetrics() ([]prometheus.Collector, []*QueryInfo)
}
