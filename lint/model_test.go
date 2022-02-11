package lint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestParseDashboard(t *testing.T) {
	sampleDashboard, err := ioutil.ReadFile("testdata/dashboard.json")
	assert.NoError(t, err)
	t.Run("Row panels", func(t *testing.T) {
		dashboard, err := NewDashboard(sampleDashboard)
		assert.NoError(t, err)
		assert.Len(t, dashboard.GetPanels(), 4)
	})
}
