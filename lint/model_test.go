package lint

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDatasource(t *testing.T) {
	for _, tc := range []struct {
		input    []byte
		expected Datasource
		err      error
	}{
		{
			input:    []byte(`"${datasource}"`),
			expected: "${datasource}",
		},
		{
			input:    []byte(`{"uid":"${datasource}"}`),
			expected: "${datasource}",
		},
		{
			input: []byte(`1`),
			err:   fmt.Errorf("invalid type for field 'datasource': 1"),
		},
		{
			input: []byte(`{}`),
			err:   fmt.Errorf("invalid type for field 'datasource': missing uid field"),
		},
		{
			input: []byte(`{"uid":1}`),
			err:   fmt.Errorf("invalid type for field 'datasource': invalid uid field type, should be string"),
		},
	} {
		var actual Datasource
		err := json.Unmarshal(tc.input, &actual)
		require.Equal(t, tc.err, err)
		require.Equal(t, tc.expected, actual)
	}
}
