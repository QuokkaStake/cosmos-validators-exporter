package generators

import (
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/fetchers"
	loggerPkg "main/pkg/logger"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"cosmossdk.io/math"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
)

func TestValidatorCommissionRateGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.State{}
	generator := NewValidatorCommissionRateGenerator([]*config.Chain{}, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestValidatorCommissionRateGeneratorNoConsumerCommissions(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{Name: "chain"}}
	state := statePkg.State{}
	state.Set(constants.FetcherNameConsumerCommission, fetchers.ConsumerCommissionData{})
	generator := NewValidatorCommissionRateGenerator(chains, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestValidatorCommissionRateInvalidConsumer(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:      "chain",
		BaseDenom: "uatom",
		Validators: []config.Validator{
			{Address: "test"},
		},
	}}
	state := statePkg.State{}
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{
		Validators: map[string]*types.ValidatorsResponse{
			"chain": {
				Validators: []types.Validator{
					{OperatorAddress: "anothertest"},
				},
			},
		},
	})
	state.Set(constants.FetcherNameConsumerCommission, fetchers.ConsumerCommissionData{})
	generator := NewValidatorCommissionRateGenerator(chains, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.Len(t, results, 1)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 0, testutil.CollectAndCount(gauge))
}

func TestValidatorCommissionRateGeneratorSuccess(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:      "chain",
		BaseDenom: "uatom",
		Validators: []config.Validator{
			{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"},
			{Address: "cosmosvaloper1qaa9zej9a0ge3ugpx3pxyx602lxh3ztqgfnp42"},
		},
		ConsumerChains: []*config.ConsumerChain{
			{Name: "consumer"},
			{Name: "otherconsumer"},
			{Name: "anotherconsumer"},
		},
	}, {Name: "otherchain"}}
	state := statePkg.State{}
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{
		Validators: map[string]*types.ValidatorsResponse{
			"chain": {
				Validators: []types.Validator{
					{
						OperatorAddress: "cosmosvaloper1c4k24jzduc365kywrsvf5ujz4ya6mwympnc4en",
					},
					{
						OperatorAddress: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
						Status:          constants.ValidatorStatusBonded,
						Commission: types.ValidatorCommission{
							CommissionRates: types.ValidatorCommissionRates{
								Rate: math.LegacyMustNewDecFromStr("0.2"),
							},
						},
					},
					{
						OperatorAddress: "cosmosvaloper14lultfckehtszvzw4ehu0apvsr77afvyju5zzy",
					},
				},
			},
		},
	})
	state.Set(constants.FetcherNameConsumerCommission, fetchers.ConsumerCommissionData{
		Commissions: map[string]map[string]*types.ConsumerCommissionResponse{
			"consumer": {
				"cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e": {Rate: math.LegacyMustNewDecFromStr("0.1")},
			},
			"otherconsumer": {},
		},
	})
	generator := NewValidatorCommissionRateGenerator(chains, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.Len(t, results, 1)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 2, testutil.CollectAndCount(gauge))
	assert.InDelta(t, 0.1, testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain":   "consumer",
		"address": "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
	})), 0.01)
	assert.InDelta(t, 0.2, testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain":   "chain",
		"address": "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
	})), 0.01)
}
