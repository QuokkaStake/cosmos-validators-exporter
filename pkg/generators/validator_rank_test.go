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

func TestValidatorRankGeneratorNoState(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewValidatorRankGenerator([]*config.Chain{}, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.Empty(t, results)
}

func TestValidatorRankGeneratorNoChainValidators(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{Name: "chain"}}
	state := statePkg.NewState()
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{})
	generator := NewValidatorRankGenerator(chains, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 0, testutil.CollectAndCount(gauge))
}

func TestValidatorRankGeneratorNotFound(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:       "chain",
		Validators: []config.Validator{{Address: "validator"}},
	}}
	state := statePkg.NewState()
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{
		Validators: map[string]*types.ValidatorsResponse{
			"chain": {
				Validators: []types.Validator{
					{Tokens: math.LegacyMustNewDecFromStr("2"), OperatorAddress: "first"},
					{Tokens: math.LegacyMustNewDecFromStr("1"), OperatorAddress: "second"},
					{Tokens: math.LegacyMustNewDecFromStr("3"), OperatorAddress: "third"},
				},
			},
		},
	})
	generator := NewValidatorRankGenerator(chains, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 0, testutil.CollectAndCount(gauge))
}

func TestValidatorRankGeneratorNotActive(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:       "chain",
		Validators: []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
	}}
	state := statePkg.NewState()
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{
		Validators: map[string]*types.ValidatorsResponse{
			"chain": {
				Validators: []types.Validator{
					{
						Tokens:          math.LegacyMustNewDecFromStr("2"),
						OperatorAddress: "cosmosvaloper1c4k24jzduc365kywrsvf5ujz4ya6mwympnc4en",
					},
					{
						Tokens:          math.LegacyMustNewDecFromStr("1"),
						OperatorAddress: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
					},
					{
						Tokens:          math.LegacyMustNewDecFromStr("3"),
						OperatorAddress: "cosmosvaloper14lultfckehtszvzw4ehu0apvsr77afvyju5zzy",
					},
				},
			},
		},
	})
	generator := NewValidatorRankGenerator(chains, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 0, testutil.CollectAndCount(gauge))
}

func TestValidatorRankGeneratorActive(t *testing.T) {
	t.Parallel()

	chains := []*config.Chain{{
		Name:       "chain",
		Validators: []config.Validator{{Address: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e"}},
	}}
	state := statePkg.NewState()
	state.Set(constants.FetcherNameValidators, fetchers.ValidatorsData{
		Validators: map[string]*types.ValidatorsResponse{
			"chain": {
				Validators: []types.Validator{
					{
						Tokens:          math.LegacyMustNewDecFromStr("3"),
						OperatorAddress: "cosmosvaloper1c4k24jzduc365kywrsvf5ujz4ya6mwympnc4en",
						Status:          constants.ValidatorStatusBonded,
					},
					{
						Tokens:          math.LegacyMustNewDecFromStr("2"),
						OperatorAddress: "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
						Status:          constants.ValidatorStatusBonded,
					},
					{
						Tokens:          math.LegacyMustNewDecFromStr("1"),
						OperatorAddress: "cosmosvaloper14lultfckehtszvzw4ehu0apvsr77afvyju5zzy",
						Status:          constants.ValidatorStatusBonded,
					},
				},
			},
		},
	})
	generator := NewValidatorRankGenerator(chains, loggerPkg.GetNopLogger())
	results := generator.Generate(state)
	assert.NotEmpty(t, results)

	gauge, ok := results[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.InEpsilon(t, float64(2), testutil.ToFloat64(gauge.With(prometheus.Labels{
		"chain":   "chain",
		"address": "cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
	})), 0.01)
}
