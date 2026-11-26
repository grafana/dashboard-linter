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
			result: ResultSuccess,
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
			result: ResultSuccess,
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
			result: ResultSuccess,
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
			result: ResultSuccess,
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

func TestPanelDatasourceBuiltin(t *testing.T) {
	linter := NewPanelDatasourceRule()

	templated := []Template{{Type: "datasource", Name: "datasource"}}

	for _, tc := range []struct {
		name       string
		datasource interface{}
		result     Result
	}{
		{
			name:       "dashboard datasource as a string",
			datasource: "-- Dashboard --",
			result:     ResultSuccess,
		},
		{
			name:       "mixed datasource as a string",
			datasource: "-- Mixed --",
			result:     ResultSuccess,
		},
		{
			name:       "grafana datasource as a string",
			datasource: "-- Grafana --",
			result:     ResultSuccess,
		},
		{
			name:       "mixed datasource as an object",
			datasource: map[string]interface{}{"type": "datasource", "uid": "-- Mixed --"},
			result:     ResultSuccess,
		},
		{
			name:       "dashboard datasource as an object",
			datasource: map[string]interface{}{"type": "datasource", "uid": "-- Dashboard --"},
			result:     ResultSuccess,
		},
		{
			name:       "grafana datasource as an object",
			datasource: map[string]interface{}{"type": "datasource", "uid": "grafana"},
			result:     ResultSuccess,
		},
		{
			name:       "mixed datasource without a type",
			datasource: map[string]interface{}{"uid": "-- Mixed --"},
			result:     ResultSuccess,
		},
		{
			name:       "templated datasource as an object",
			datasource: map[string]interface{}{"type": "prometheus", "uid": "${datasource}"},
			result:     ResultSuccess,
		},
		{
			name:       "literal uid with a plugin type",
			datasource: map[string]interface{}{"type": "prometheus", "uid": "pjNv0fYGz"},
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test', panel 'bar' does not use a templated datasource, uses 'pjNv0fYGz'",
			},
		},
		{
			name:       "datasource that only looks like a built-in one",
			datasource: "Dashboard",
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test', panel 'bar' does not use a templated datasource, uses 'Dashboard'",
			},
		},
	} {
		for _, panelType := range []string{"singlestat", "graph", "table", "timeseries"} {
			t.Run(tc.name+"/"+panelType, func(t *testing.T) {
				testRule(t, linter, Dashboard{
					Title: "test",
					Panels: []Panel{{
						Type:       panelType,
						Title:      "bar",
						Datasource: tc.datasource,
					}},
					Templating: struct {
						List []Template "json:\"list\""
					}{List: templated},
				}, tc.result)
			})
		}
	}
}

// testRule is a small helper that tests a lint rule and expects it to only return
// a single result.
func testRule(t *testing.T, rule Rule, d Dashboard, result Result) {
	testRuleWithAutofix(t, rule, &d, []Result{result}, false)
}
func testMultiResultRule(t *testing.T, rule Rule, d Dashboard, result []Result) {
	testRuleWithAutofix(t, rule, &d, result, false)
}

func testRuleWithAutofix(t *testing.T, rule Rule, d *Dashboard, result []Result, autofix bool) {
	rs := ResultSet{}
	rule.Lint(*d, &rs)
	if autofix {
		rs.AutoFix(d)
	}
	require.Len(t, rs.results, 1)
	actual := rs.results[0].Result
	if actual.Results[0].Severity == Quiet {
		// all test cases expect success
		actual.Results[0].Severity = Success
	}
	rr := make([]Result, len(actual.Results))
	for i, r := range actual.Results {
		rr[i] = r.Result
	}

	require.Equal(t, result, rr)
}
