package lint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPanelDatasource(t *testing.T) {
	linter := NewPanelDatasourceRule()

	for _, tc := range []struct {
		result Result
		panel  Panel
	}{
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test', panel 'bar' does not use $datasource for datasource, uses 'foo'",
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
		},
	} {
		testRule(t, linter, Dashboard{Title: "test",
			Templating: struct {
				List []Template `json:"list"`
			}{
				List: []Template{
					{
						Type:  "datasource",
						Query: "prometheus",
						Name:  "datasource",
					},
				},
			},

			Panels: []Panel{tc.panel}}, tc.result)
	}
}

func TestPanelMultiDatasource(t *testing.T) {
	linter := NewPanelDatasourceRule()

	for _, tc := range []struct {
		result Result
		panel  Panel
	}{
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test', panel 'bar' does not use $prometheus_datasource or $loki_datasource for datasource, uses '$datasource'",
			},
			panel: Panel{
				Type:       "singlestat",
				Datasource: "$datasource",
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
				Datasource: "$prometheus_datasource",
			},
		},
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			panel: Panel{
				Type:       "singlestat",
				Datasource: "$loki_datasource",
			},
		},
	} {
		testRule(t, linter, Dashboard{Title: "test",
			Templating: struct {
				List []Template `json:"list"`
			}{
				List: []Template{
					{
						Type:  "datasource",
						Query: "prometheus",
						Name:  "datasource",
					},
					{
						Type:  "datasource",
						Query: "loki",
						Name:  "datasource",
					},
				},
			},

			Panels: []Panel{tc.panel}}, tc.result)
	}
}

// testRule is a small helper that tests a lint rule and expects it to only return
// a single result.
func testRule(t *testing.T, rule Rule, d Dashboard, result Result) {
	var rs ResultSet
	rule.Lint(d, &rs)
	require.Len(t, rs.results, 1)
	require.Equal(t, result, rs.results[0].Result)
}
