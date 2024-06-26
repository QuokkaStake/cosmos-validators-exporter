package generators

import (
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
)

func TestConsumerNeedsToSignGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewConsumerNeedsToSignGenerator([]*config.Chain{})
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestConsumerNeedsToSignGeneratorNotEmptyState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameValidatorConsumers, fetchers.ValidatorConsumersData{
		Infos: map[string]map[string]map[string]bool{
			"provider": {
				"validator": map[string]bool{
					"consumer-chain-id": true,
				},
			},
		},
	})

	chains := []*config.Chain{{
		Name:       "provider",
		Validators: []config.Validator{{Address: "validator"}, {Address: "othervalidator"}},
		ConsumerChains: []*config.ConsumerChain{{
			Name:    "consumer",
			ChainID: "consumer-chain-id",
		}, {
			Name:    "otherconsumer",
			ChainID: "otherconsumer-chain-id",
		}},
	}, {Name: "otherprovider"}}

	generator := NewConsumerNeedsToSignGenerator(chains)
	results := generator.Generate(state)
	assert.Len(t, results, 1)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 2, testutil.CollectAndCount(gauge))
	assert.InEpsilon(t, float64(1), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain":    "consumer",
		"chain_id": "consumer-chain-id",
		"provider": "provider",
		"address":  "validator",
	})), 0.01)
	assert.Zero(t, testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain":    "otherconsumer",
		"chain_id": "otherconsumer-chain-id",
		"provider": "provider",
		"address":  "validator",
	})))
}
