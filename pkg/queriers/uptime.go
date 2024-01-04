package queriers

import (
	"main/pkg/types"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type UptimeQuerier struct {
	StartTime time.Time
}

func NewUptimeQuerier() *UptimeQuerier {
	return &UptimeQuerier{
		StartTime: time.Now(),
	}
}

func (u *UptimeQuerier) GetMetrics() ([]prometheus.Collector, []*types.QueryInfo) {
	uptimeMetricsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_start_time",
			Help: "Unix timestamp on when the app was started. Useful for annotations.",
		},
		[]string{},
	)

	uptimeMetricsGauge.With(prometheus.Labels{}).Set(float64(u.StartTime.Unix()))
	return []prometheus.Collector{uptimeMetricsGauge}, []*types.QueryInfo{}
}
