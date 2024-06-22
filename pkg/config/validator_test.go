package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorDisplayWarningNoConsensusAddress(t *testing.T) {
	t.Parallel()

	validator := Validator{Address: "test"}
	warnings := validator.DisplayWarnings(&Chain{Name: "test"})
	assert.NotEmpty(t, warnings)
}

func TestValidatorDisplayWarningEmpty(t *testing.T) {
	t.Parallel()

	validator := Validator{Address: "test", ConsensusAddress: "test"}
	warnings := validator.DisplayWarnings(&Chain{Name: "test"})
	assert.Empty(t, warnings)
}
