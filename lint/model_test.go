package lint

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDatasource(t *testing.T) {
	for _, tc := range []struct {
		name     string
		input    []byte
		expected Datasource
		err      error
	}{
		{
			name:     "string",
			input:    []byte(`"${datasource}"`),
			expected: Datasource{"${datasource}", ""},
		},
		{
			name:     "uid",
			input:    []byte(`{"uid":"${datasource}"}`),
			expected: Datasource{"${datasource}", ""},
		},
		{
			name:     "uid-type",
			input:    []byte(`{"uid":"${datasource}","type":"${type}"}`),
			expected: Datasource{"${datasource}", "${type}"},
		},
		{
			name:  "byte",
			input: []byte(`1`),
			err:   fmt.Errorf("invalid type for field 'datasource': 1"),
		},
		{
			name:  "empty object",
			input: []byte(`{}`),
			err:   fmt.Errorf("invalid type for field 'datasource': missing uid field"),
		},
		{
			name:  "int uid",
			input: []byte(`{"uid":1}`),
			err:   fmt.Errorf("invalid type for field 'datasource': invalid uid field type, should be string"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var raw interface{}
			err := json.Unmarshal(tc.input, &raw)
			require.NoError(t, err)
			actual, err := GetDataSource(raw)
			require.Equal(t, tc.err, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestParseDashboard(t *testing.T) {
	sampleDashboard, err := os.ReadFile("testdata/dashboard.json")
	assert.NoError(t, err)
	t.Run("Row panels", func(t *testing.T) {
		dashboard, err := NewDashboard(sampleDashboard)
		assert.NoError(t, err)
		assert.Len(t, dashboard.GetPanels(), 4)
	})
	t.Run("Annotations", func(t *testing.T) {
		dashboard, err := NewDashboard(sampleDashboard)
		assert.NoError(t, err)
		assert.Len(t, dashboard.Annotations.List, 1)
	})

	t.Run("v0alpha1 dashboard", func(t *testing.T) {
		wrap := `{
			"apiVersion": "v0alpha1",
			"kind": "Dashboard",
			"spec": ` + string(sampleDashboard) + `
		}`

		dashboard, err := NewDashboard([]byte(wrap))
		assert.NoError(t, err)
		assert.Len(t, dashboard.Annotations.List, 1)
		assert.Equal(t, "v0alpha1", dashboard.APIVersion)
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
			expected: TemplateValue{Text: "", Value: ""},
		},
		{
			input:    []byte(`{"text": "text"}`),
			expected: TemplateValue{Text: "text", Value: ""},
		},
	} {
		var raw RawTemplateValue
		err := json.Unmarshal(tc.input, &raw)
		require.NoError(t, err)
		actual, err := raw.Get()
		require.Equal(t, tc.err, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestParseTemplate(t *testing.T) {
	for _, tc := range []struct {
		input    []byte
		expected Template
		err      error
	}{
		{
			// NB no "query.query" field, some data source don't use this.
			input:    []byte(`{ "type": "query", "query": {} }`),
			expected: Template{Type: "query", RawQuery: map[string]interface{}{}},
		},
	} {
		var actual Template
		err := json.Unmarshal(tc.input, &actual)
		require.NoError(t, err)
		require.Equal(t, tc.expected, actual)
	}
}
