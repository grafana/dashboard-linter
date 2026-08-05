package lint

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// v2MinimalSpec contains all required fields — the rule must pass.
const v2MinimalSpec = `{
	"apiVersion": "dashboard.grafana.app/v2",
	"kind": "Dashboard",
	"spec": {
		"title": "Minimal",
		"cursorSync": "Off",
		"preload": false,
		"annotations": [],
		"elements": {},
		"layout": {"kind": "GridLayout", "spec": {"items": []}},
		"links": [],
		"tags": [],
		"timeSettings": {"from": "now-6h", "to": "now"},
		"variables": []
	}
}`

func TestV2RequiredFieldsRule_AllPresent(t *testing.T) {
	d, err := NewDashboard([]byte(v2MinimalSpec))
	require.NoError(t, err)
	rule := NewV2RequiredFieldsRule()
	results := rule.fn(d)
	assert.Empty(t, results.Results, "all required fields present — rule should pass")
}

func TestV2RequiredFieldsRule_MissingFields(t *testing.T) {
	for _, f := range v2RequiredFields {
		f := f
		t.Run("missing_"+f.name, func(t *testing.T) {
			trimmed := strings.Replace(v2MinimalSpec, `"`+f.name+`"`, `"_removed_`+f.name+`"`, 1)
			d, err := NewDashboard([]byte(trimmed))
			require.NoError(t, err)
			rule := NewV2RequiredFieldsRule()
			results := rule.fn(d)
			require.NotEmpty(t, results.Results, "expected an error for missing %q", f.name)
			assert.Equal(t, Error, results.Results[0].Severity)
		})
	}
}

func TestV2RequiredFieldsRule_WrongType(t *testing.T) {
	// Replace "annotations": [] with "annotations": {} to trigger a type error.
	wrongType := strings.Replace(v2MinimalSpec, `"annotations": []`, `"annotations": {}`, 1)
	// NewDashboard must succeed (not abort) even though the type is wrong.
	d, err := NewDashboard([]byte(wrongType))
	require.NoError(t, err)
	rule := NewV2RequiredFieldsRule()
	results := rule.fn(d)
	require.NotEmpty(t, results.Results)
	assert.Equal(t, Error, results.Results[0].Severity)
	assert.Contains(t, results.Results[0].Message, "annotations")
}

func TestV2RequiredFieldsRule_NonV2Skipped(t *testing.T) {
	classic := `{"title": "Classic", "panels": []}`
	d, err := NewDashboard([]byte(classic))
	require.NoError(t, err)
	rule := NewV2RequiredFieldsRule()
	results := rule.fn(d)
	assert.Empty(t, results.Results, "classic dashboard must not be checked by v2 rule")
}

func TestV2RequiredFieldsRule_EndToEnd(t *testing.T) {
	noTitle := strings.Replace(v2Dashboard, `"title": "V2 Test",`, `"_title": "V2 Test",`, 1)
	d, err := NewDashboard([]byte(noTitle))
	require.NoError(t, err)
	rs := NewRuleSet()
	results, err := rs.Lint([]Dashboard{d})
	require.NoError(t, err)
	assert.True(t, ruleHasError(results, "v2-required-fields-rule"),
		"end-to-end: rule must fire when title is absent")
}

func TestV2RequiredFieldsRule_WrongTypeEndToEnd(t *testing.T) {
	// Simulate the exact error from the bug report: annotations is an object, not array.
	wrongAnnotations := fmt.Sprintf(`{
		"apiVersion": "dashboard.grafana.app/v2",
		"kind": "Dashboard",
		"spec": {
			"title": "Bad",
			"cursorSync": "Off",
			"preload": false,
			"annotations": {"builtin": 0},
			"elements": {},
			"layout": {"kind": "GridLayout", "spec": {"items": []}},
			"links": [],
			"tags": [],
			"timeSettings": {"from": "now-6h", "to": "now"},
			"variables": []
		}
	}`)
	// Must not return a parse error — degraded dashboard returned instead.
	d, err := NewDashboard([]byte(wrongAnnotations))
	require.NoError(t, err)
	rs := NewRuleSet()
	results, err := rs.Lint([]Dashboard{d})
	require.NoError(t, err)
	assert.True(t, ruleHasError(results, "v2-required-fields-rule"),
		"wrong type for annotations must be reported by the rule, not abort parsing")
}
