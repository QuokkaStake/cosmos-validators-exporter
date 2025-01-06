package generators

import (
	"github.com/prometheus/client_golang/prometheus"
	"main/pkg/fetchers"
)

type Generator interface {
	Generate(state fetchers.State) []prometheus.Collector
}
