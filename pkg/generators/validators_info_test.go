package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"cosmossdk.io/math"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
)

func TestValidatorsInfoGeneratorNoValidators(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewValidatorsInfoGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestValidatorsInfoGeneratorNoConsumerValidators(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{})
	generator := NewValidatorsInfoGenerator()
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestValidatorsInfoGeneratorNotConsumer(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{
		Validators: map[string]*types.ValidatorsResponse{
			"chain": {
				Validators: []types.Validator{
					{
						DelegatorShares: math.LegacyMustNewDecFromStr("2"),
						OperatorAddress: "cosmosvaloper1c4k24jzduc365kywrsvf5ujz4ya6mwympnc4en",
						Status:          constants.ValidatorStatusBonded,
					},
					{
						DelegatorShares: math.LegacyMustNewDecFromStr("1"),
						OperatorAddress: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
					},
					{
						DelegatorShares: math.LegacyMustNewDecFromStr("3"),
						OperatorAddress: "cosmosvaloper14lultfckehtszvzw4ehu0apvsr77afvyju5zzy",
						Status:          constants.ValidatorStatusBonded,
					},
				},
			},
		},
	})
	state.Set(constants.FetcherNameConsumerValidators, fetchers.ConsumerValidatorsData{})
	generator := NewValidatorsInfoGenerator()
	results := generator.Generate(state)
	assert.Len(t, results, 2)

	validatorsCountGauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(2), testutil.ToFloat64(validatorsCountGauge.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)

	totalBondedGauge, ok := results[1].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(5), testutil.ToFloat64(totalBondedGauge.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)
}

func TestValidatorsInfoGeneratorConsumer(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{})
	state.Set(constants.FetcherNameConsumerValidators, fetchers.ConsumerValidatorsData{
		Validators: map[string]*types.ConsumerValidatorsResponse{
			"chain": {
				Validators: []types.ConsumerValidator{
					{},
					{},
					{},
				},
			},
		},
	})
	generator := NewValidatorsInfoGenerator()
	results := generator.Generate(state)
	assert.Len(t, results, 2)

	validatorsCountGauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(3), testutil.ToFloat64(validatorsCountGauge.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)

	totalBondedGauge, ok := results[1].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 0, testutil.CollectAndCount(totalBondedGauge))
}
