package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChainMethods(t *testing.T) {
	t.Parallel()

	chain := Chain{
		LCDEndpoint: "example",
		Name:        "chain",
		Queries:     map[string]bool{"enabled": true},
	}

	assert.Equal(t, "example", chain.GetHost())
	assert.Equal(t, "chain", chain.GetName())
	assert.Len(t, chain.GetQueries(), 1)
}

func TestChainValidateNoName(t *testing.T) {
	t.Parallel()

	chain := Chain{}
	err := chain.Validate()
	require.Error(t, err)
}

func TestChainValidateNoEndpoint(t *testing.T) {
	t.Parallel()

	chain := Chain{Name: "test"}
	err := chain.Validate()
	require.Error(t, err)
}

func TestChainValidateNoValidators(t *testing.T) {
	t.Parallel()

	chain := Chain{Name: "test", LCDEndpoint: "test"}
	err := chain.Validate()
	require.Error(t, err)
}

func TestChainValidateInvalidValidator(t *testing.T) {
	t.Parallel()

	chain := Chain{
		Name:        "test",
		LCDEndpoint: "test",
		Validators:  []Validator{{}},
	}
	err := chain.Validate()
	require.Error(t, err)
}

func TestChainValidateInvalidDenom(t *testing.T) {
	t.Parallel()

	chain := Chain{
		Name:        "test",
		LCDEndpoint: "test",
		Validators:  []Validator{{Address: "test"}},
		Denoms:      DenomInfos{{}},
	}
	err := chain.Validate()
	require.Error(t, err)
}

func TestChainValidateInvalidConsumer(t *testing.T) {
	t.Parallel()

	chain := Chain{
		Name:           "test",
		LCDEndpoint:    "test",
		Validators:     []Validator{{Address: "test"}},
		Denoms:         DenomInfos{{Denom: "ustake", DisplayDenom: "stake"}},
		ConsumerChains: []*ConsumerChain{{}},
	}
	err := chain.Validate()
	require.Error(t, err)
}

func TestChainValidateValid(t *testing.T) {
	t.Parallel()

	chain := Chain{
		Name:        "test",
		LCDEndpoint: "test",
		Validators:  []Validator{{Address: "test"}},
		Denoms:      DenomInfos{{Denom: "ustake", DisplayDenom: "stake"}},
	}
	err := chain.Validate()
	require.NoError(t, err)
}

func TestChainDisplayWarningsNoBaseDenom(t *testing.T) {
	t.Parallel()

	chain := Chain{
		Name:        "test",
		LCDEndpoint: "test",
		Validators:  []Validator{{Address: "test"}},
		Denoms:      DenomInfos{{Denom: "ustake", DisplayDenom: "stake", CoingeckoCurrency: "stake"}},
	}
	warnings := chain.DisplayWarnings()
	require.NotEmpty(t, warnings)
}

func TestChainDisplayWarningsDenomWarning(t *testing.T) {
	t.Parallel()

	chain := Chain{
		Name:        "test",
		LCDEndpoint: "test",
		BaseDenom:   "ustake",
		Validators:  []Validator{{Address: "test"}},
		Denoms:      DenomInfos{{Denom: "ustake", DisplayDenom: "stake"}},
	}
	warnings := chain.DisplayWarnings()
	require.NotEmpty(t, warnings)
}

func TestChainDisplayWarningsEmpty(t *testing.T) {
	t.Parallel()

	chain := Chain{
		Name:        "test",
		LCDEndpoint: "test",
		BaseDenom:   "ustake",
		Validators:  []Validator{{Address: "test"}},
		Denoms:      DenomInfos{{Denom: "ustake", DisplayDenom: "stake", CoingeckoCurrency: "stake"}},
	}
	warnings := chain.DisplayWarnings()
	require.Empty(t, warnings)
}