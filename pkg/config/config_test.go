package config

import (
	"testing"

	"github.com/guregu/null/v5"
	"github.com/stretchr/testify/require"
)

func TestConfigValidateInvalidTracingConfig(t *testing.T) {
	t.Parallel()

	config := Config{
		TracingConfig: TracingConfig{Enabled: null.BoolFrom(true)},
	}

	err := config.Validate()
	require.Error(t, err)
}

func TestConfigValidateNoChains(t *testing.T) {
	t.Parallel()

	config := Config{}

	err := config.Validate()
	require.Error(t, err)
}

func TestConfigValidateInvalidChain(t *testing.T) {
	t.Parallel()

	config := Config{
		Chains: []*Chain{{}},
	}

	err := config.Validate()
	require.Error(t, err)
}

func TestConfigValidateValid(t *testing.T) {
	t.Parallel()

	config := Config{
		Chains: []*Chain{{
			Name:        "chain",
			LCDEndpoint: "test",
			Validators:  []Validator{{Address: "test"}},
		}},
	}

	err := config.Validate()
	require.NoError(t, err)
}

func TestDisplayWarningsChainWarning(t *testing.T) {
	t.Parallel()

	config := Config{
		Chains: []*Chain{{
			Name:        "chain",
			LCDEndpoint: "test",
			Validators:  []Validator{{Address: "test"}},
		}},
	}

	warnings := config.DisplayWarnings()
	require.NotEmpty(t, warnings)
}

func TestDisplayWarningsEmpty(t *testing.T) {
	t.Parallel()

	config := Config{
		Chains: []*Chain{{
			Name:        "chain",
			LCDEndpoint: "test",
			BaseDenom:   "test",
			Validators:  []Validator{{Address: "test"}},
		}},
	}

	warnings := config.DisplayWarnings()
	require.Empty(t, warnings)
}

func TestCoingeckoCurrencies(t *testing.T) {
	t.Parallel()

	config := Config{
		Chains: []*Chain{{
			Denoms: DenomInfos{
				{Denom: "denom1", CoingeckoCurrency: "denom1"},
				{Denom: "denom2"},
			},
			ConsumerChains: []*ConsumerChain{{
				Denoms: DenomInfos{
					{Denom: "denom3", CoingeckoCurrency: "denom3"},
					{Denom: "denom4"},
				},
			}},
		}},
	}

	currencies := config.GetCoingeckoCurrencies()
	require.Len(t, currencies, 2)
	require.Contains(t, currencies, "denom1")
	require.Contains(t, currencies, "denom3")
}
