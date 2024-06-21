package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestChainQueryEnabled(t *testing.T) {
	t.Parallel()

	queries := Queries{"query1": true, "query2": false}
	assert.True(t, queries.Enabled("query1"))
	assert.False(t, queries.Enabled("query2"))
	assert.True(t, queries.Enabled("query3"))
}
