package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsumerChainMethods(t *testing.T) {
	t.Parallel()

	chain := ConsumerChain{
		LCDEndpoint: "example",
		Name:        "chain",
		Queries:     map[string]bool{"enabled": true},
	}

	assert.Equal(t, "example", chain.GetHost())
	assert.Equal(t, "chain", chain.GetName())
	assert.Len(t, chain.GetQueries(), 1)
}

func TestConsumerChainValidateNoName(t *testing.T) {
	t.Parallel()

	chain := ConsumerChain{}
	err := chain.Validate()
	require.Error(t, err)
}

func TestConsumerChainValidateNoEndpoint(t *testing.T) {
	t.Parallel()

	chain := ConsumerChain{Name: "test"}
	err := chain.Validate()
	require.Error(t, err)
}

func TestConsumerChainValidateNoChainId(t *testing.T) {
	t.Parallel()

	chain := ConsumerChain{Name: "test", LCDEndpoint: "test"}
	err := chain.Validate()
	require.Error(t, err)
}

func TestConsumerChainValidateNoBaseDenom(t *testing.T) {
	t.Parallel()

	chain := ConsumerChain{Name: "test", LCDEndpoint: "test", ConsumerID: "0"}
	err := chain.Validate()
	require.Error(t, err)
}

func TestConsumerChainValidateInvalidDenom(t *testing.T) {
	t.Parallel()

	chain := ConsumerChain{
		Name:        "test",
		LCDEndpoint: "test",
		ConsumerID:  "0",
		BaseDenom:   "denom",
		Denoms:      DenomInfos{{}},
	}
	err := chain.Validate()
	require.Error(t, err)
}

func TestConsumerChainValidateValid(t *testing.T) {
	t.Parallel()

	chain := ConsumerChain{
		Name:        "test",
		LCDEndpoint: "test",
		ConsumerID:  "0",
		BaseDenom:   "denom",
		Denoms:      DenomInfos{{Denom: "ustake", DisplayDenom: "stake"}},
	}
	err := chain.Validate()
	require.NoError(t, err)
}
