package generators

import (
	"main/pkg/config"
	statePkg "main/pkg/state"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsConsumerGenerator(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	chains := []*config.Chain{
		{
			Name: "chain",
			ConsumerChains: []*config.ConsumerChain{
				{
					Name: "consumer",
				},
			},
		},
	}

	generator := NewIsConsumerGenerator(chains)
	results := generator.Generate(state)
	assert.NotEmpty(t, results)
}
