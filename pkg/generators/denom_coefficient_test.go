package generators

import (
	"main/pkg/config"
	statePkg "main/pkg/state"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDenomCoefficientGenerator(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	chains := []*config.Chain{
		{
			Name:      "chain",
			BaseDenom: "uatom",
			Denoms: config.DenomInfos{
				{Denom: "uatom", DisplayDenom: "atom", DenomCoefficient: 1000000},
			},
			ConsumerChains: []*config.ConsumerChain{
				{
					Name:      "consumer",
					BaseDenom: "uakt",
					Denoms: config.DenomInfos{
						{Denom: "uakt", DisplayDenom: "akt", DenomCoefficient: 1000000},
					},
				},
			},
		},
	}

	generator := NewDenomCoefficientGenerator(chains)
	results := generator.Generate(state)
	assert.NotEmpty(t, results)
}
