package tendermint

import (
	configPkg "main/pkg/config"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type RPCWithConsumers struct {
	RPC       *RPC
	Consumers []*RPC
}

func RPCWithConsumersFromChain(
	chain *configPkg.Chain,
	timeout int,
	logger zerolog.Logger,
	tracer trace.Tracer,
) *RPCWithConsumers {
	consumers := make([]*RPC, len(chain.ConsumerChains))

	for index, consumer := range chain.ConsumerChains {
		consumers[index] = NewRPC(consumer, timeout, logger, tracer)
	}

	return &RPCWithConsumers{
		Consumers: consumers,
		RPC:       NewRPC(chain, timeout, logger, tracer),
	}
}
