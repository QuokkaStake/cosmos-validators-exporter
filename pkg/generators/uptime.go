package generators

import (
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type UptimeGenerator struct {
	StartTime time.Time
}

func NewUptimeGenerator() *UptimeGenerator {
	return &UptimeGenerator{StartTime: time.Now()}
}

func (g *UptimeGenerator) Generate(state fetchersPkg.State) []prometheus.Collector {
	uptimeMetricsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "start_time",
			Help: "Unix timestamp on when the app was started. Useful for annotations.",
		},
		[]string{},
	)

	uptimeMetricsGauge.With(prometheus.Labels{}).Set(float64(g.StartTime.Unix()))

	return []prometheus.Collector{uptimeMetricsGauge}
}
