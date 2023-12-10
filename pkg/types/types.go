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

type Amount struct {
	Amount float64
	Denom  string
}

type Querier interface {
	GetMetrics() ([]prometheus.Collector, []*QueryInfo)
}
