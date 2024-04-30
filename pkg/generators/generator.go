package generators

import (
	"github.com/prometheus/client_golang/prometheus"
	statePkg "main/pkg/state"
)

type Generator interface {
	Generate(state *statePkg.State) []prometheus.Collector
}
