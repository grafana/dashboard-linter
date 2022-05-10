package lint

import (
	"encoding/json"
	"errors"
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

func TestParseTemplateValue(t *testing.T) {
	for _, tc := range []struct {
		input    []byte
		expected TemplateValue
		err      error
	}{
		{
			input:    []byte(`{"text": "text", "value": "value"}`),
			expected: TemplateValue{Text: "text", Value: "value"},
		},
		{
			input:    []byte(`{"text": ["text1", "text2"], "value": ["value1", "value2"]}`),
			expected: TemplateValue{Text: "text1", Value: "value1"},
		},
		{
			input: []byte(`{"text": 1, "value": 2}`),
			err:   errors.New("invalid type for field 'text': 1"),
		},
		{
			input:    []byte(`{"text": "text", "value": 2}`),
			expected: TemplateValue{Text: "text"},
			err:      errors.New("invalid type for field 'value': 2"),
		},
		{
			input:    []byte(`{}`),
			expected: TemplateValue{Text:"", Value:""},
		},
		{
			input:    []byte(`{"text": "text"}`),
			expected: TemplateValue{Text:"text", Value:""},
		},
	} {
		var actual TemplateValue
		err := json.Unmarshal(tc.input, &actual)
		require.Equal(t, tc.err, err)
		require.Equal(t, tc.expected, actual)
	}
}
