package generators

import (
	statePkg "main/pkg/state"

	"github.com/prometheus/client_golang/prometheus"
)

type Generator interface {
	Generate(state *statePkg.State) []prometheus.Collector
}
