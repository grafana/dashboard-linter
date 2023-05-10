package lint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPanelDatasource(t *testing.T) {
	linter := NewPanelDatasourceRule()

	for _, tc := range []struct {
		result    Result
		panel     Panel
		templates []Template
	}{
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test', panel 'bar' does not use a templated datasource, uses 'foo'",
			},
			panel: Panel{
				Type:       "singlestat",
				Datasource: "foo",
				Title:      "bar",
			},
		},
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			panel: Panel{
				Type:       "singlestat",
				Datasource: "$datasource",
			},
			templates: []Template{
				{
					Type: "datasource",
					Name: "datasource",
				},
			},
		},
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			panel: Panel{
				Type:       "singlestat",
				Datasource: "${datasource}",
			},
			templates: []Template{
				{
					Type: "datasource",
					Name: "datasource",
				},
			},
		},
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			panel: Panel{
				Type:       "singlestat",
				Datasource: "$prometheus_datasource",
			},
			templates: []Template{
				{
					Type: "datasource",
					Name: "prometheus_datasource",
				},
			},
		},
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			panel: Panel{
				Type:       "singlestat",
				Datasource: "${prometheus_datasource}",
			},
			templates: []Template{
				{
					Type: "datasource",
					Name: "prometheus_datasource",
				},
			},
		},
	} {
		testRule(t, linter, Dashboard{
			Title:  "test",
			Panels: []Panel{tc.panel},
			Templating: struct {
				List []Template "json:\"list\""
			}{List: tc.templates},
		}, tc.result)
	}
}

// testRule is a small helper that tests a lint rule and expects it to only return
// a single result.
func testRule(t *testing.T, rule Rule, d Dashboard, result Result) {
	testRuleWithAutofix(t, rule, &d, result, false)
}

func testRuleWithAutofix(t *testing.T, rule Rule, d *Dashboard, result Result, autofix bool) {
	rs := ResultSet{
		config: &ConfigurationFile{Autofix: autofix},
	}
	rule.Lint(d, &rs)
	require.Len(t, rs.results, 1)
	actual := rs.results[0].Result
	if actual.Severity == Quiet {
		// all test cases expect success
		actual.Severity = Success
	}
	require.Equal(t, result, actual)
}
