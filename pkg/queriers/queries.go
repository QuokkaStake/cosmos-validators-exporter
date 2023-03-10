package queriers

import (
	"main/pkg/config"
	"main/pkg/types"

	"github.com/prometheus/client_golang/prometheus"
)

type QueriesQuerier struct {
	Config *config.Config
	Infos  []types.QueryInfo
}

func NewQueriesQuerier(appConfig *config.Config, queryInfos []types.QueryInfo) *QueriesQuerier {
	return &QueriesQuerier{
		Config: appConfig,
		Infos:  queryInfos,
	}
}

func (q *QueriesQuerier) GetMetrics() ([]prometheus.Collector, []types.QueryInfo) {
	queriesCountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_queries_total",
			Help: "Total queries done for this chain",
		},
		[]string{"chain"},
	)

	queriesSuccessfulGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_queries_success",
			Help: "Successful queries count for this chain",
		},
		[]string{"chain"},
	)

	queriesFailedGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_queries_error",
			Help: "Failed queries count for this chain",
		},
		[]string{"chain"},
	)

	timingsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_validators_exporter_timings",
			Help: "External LCD query timing",
		},
		[]string{"chain", "url"},
	)

	// so we would have this metrics even if there are no requests
	for _, chain := range q.Config.Chains {
		queriesCountGauge.With(prometheus.Labels{
			"chain": chain.Name,
		}).Set(0)

		queriesSuccessfulGauge.With(prometheus.Labels{
			"chain": chain.Name,
		}).Set(0)

		queriesFailedGauge.With(prometheus.Labels{
			"chain": chain.Name,
		}).Set(0)
	}

	for _, query := range q.Infos {
		queriesCountGauge.With(prometheus.Labels{
			"chain": query.Chain,
		}).Inc()

		timingsGauge.With(prometheus.Labels{
			"chain": query.Chain,
			"url":   query.URL,
		}).Set(query.Duration.Seconds())

		if query.Success {
			queriesSuccessfulGauge.With(prometheus.Labels{
				"chain": query.Chain,
			}).Inc()
		} else {
			queriesFailedGauge.With(prometheus.Labels{
				"chain": query.Chain,
			}).Inc()
		}
	}

	return []prometheus.Collector{
		queriesCountGauge,
		queriesSuccessfulGauge,
		queriesFailedGauge,
		timingsGauge,
	}, []types.QueryInfo{}
}
