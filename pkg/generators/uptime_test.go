package generators

import (
	statePkg "main/pkg/state"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUptimeGenerator(t *testing.T) {
	t.Parallel()

	state := statePkg.NewState()
	generator := NewUptimeGenerator()
	results := generator.Generate(state)
	assert.NotEmpty(t, results)
}
