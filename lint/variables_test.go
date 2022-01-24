package lint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVariableExpansion(t *testing.T) {
	for _, tc := range []struct {
		desc   string
		expr   string
		result string
	}{
		{"Should not replace variables in quoted strings", "label_values(up{job=~\"$job\"}, namespace)", "label_values(up{job=~\"$job\"}, namespace)"},
	} {
		require.Equal(t, tc.result, expandVariables(tc.expr))
	}
}
