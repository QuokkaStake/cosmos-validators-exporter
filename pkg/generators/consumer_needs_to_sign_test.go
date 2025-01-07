package generators

import (
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
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

func TestConsumerNeedsToSignGeneratorNoConsumerInfo(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameValidatorConsumers, fetchers.ValidatorConsumersData{
		Infos: map[string]map[string]map[string]bool{
			"provider": {
				"validator": map[string]bool{
					"consumer-id": true,
				},
			},
		},
	})

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
					"consumer-id": true,
				},
			},
		},
	})

	state.Set(constants.FetcherNameConsumerInfo, fetchers.ConsumerInfoData{
		Info: map[string]map[string]types.ConsumerChainInfo{
			"provider": {
				"consumer-id": types.ConsumerChainInfo{
					ConsumerID: "consumer-id",
				},
			},
		},
	})

	chains := []*config.Chain{{
		Name:       "provider",
		Validators: []config.Validator{{Address: "validator"}, {Address: "othervalidator"}},
	}, {Name: "otherprovider"}}

	generator := NewConsumerNeedsToSignGenerator(chains)
	results := generator.Generate(state)
	assert.Len(t, results, 1)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 1, testutil.CollectAndCount(gauge))
	assert.InEpsilon(t, float64(1), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"consumer_id": "consumer-id",
		"provider":    "provider",
		"address":     "validator",
	})), 0.01)
}
