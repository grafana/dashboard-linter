package lint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTargetPromQLRule(t *testing.T) {
	linter := NewTargetPromQLRule()

	for _, tc := range []struct {
		result []Result
		panel  Panel
	}{
		// Don't fail non-prometheus panels.
		{
			result: []Result{ResultSuccess},
			panel: Panel{
				Title:      "panel",
				Datasource: "foo",
				Targets: []Target{
					{
						Expr: `sum(rate(foo[5m]))`,
					},
				},
			},
		},
		// This is what a valid panel looks like.
		{
			result: []Result{ResultSuccess},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate(foo[5m]))`,
					},
				},
			},
		},
		// Invalid query
		{
			result: []Result{ResultSuccess},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate(foo[5m]))`,
					},
				},
			},
		},
		// Timeseries support
		{
			result: []Result{{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'foo(bar.baz)': 1:8: parse error: unexpected character: '.'",
			}},
			panel: Panel{
				Title: "panel",
				Type:  "timeseries",
				Targets: []Target{
					{
						Expr: `foo(bar.baz)`,
					},
				},
			},
		},
		// Variable substitutions
		{
			result: []Result{ResultSuccess},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate(foo[$__rate_interval])) * $__range_s`,
					},
				},
			},
		},
		// Variable substitutions with ${...}
		{
			result: []Result{ResultSuccess},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate(foo[$__rate_interval])) * ${__range_s}`,
					},
				},
			},
		},
		// Variable substitutions inside by clause
		{
			result: []Result{ResultSuccess},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum by(${variable:csv}) (rate(foo[$__rate_interval])) * $__range_s`,
					},
				},
			},
		},
		// Template variables substitutions
		{
			result: []Result{ResultSuccess},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum (rate(foo[$interval:$resolution]))`,
					},
				},
			},
		},
		{
			result: []Result{ResultSuccess},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `increase(foo{}[$sampling])`,
					},
				},
			},
		},
		// Empty PromQL expression
		{
			result: []Result{{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query '': unknown position: parse error: no expression found in input",
			}},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: ``,
					},
				},
			},
		},
		// Reference another panel that does not exist
		{
			result: []Result{
				{
					Severity: Error,
					Message:  "Dashboard 'dashboard', panel 'panel', target idx '0' Invalid panel reference in target",
				},
				{
					Severity: Error,
					Message:  "Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query '': unknown position: parse error: no expression found in input",
				},
			},
			panel: Panel{
				Id:    1,
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						PanelId: 2,
					},
				},
			},
		},
	} {
		dashboard := Dashboard{
			Title: "dashboard",
			Templating: struct {
				List []Template `json:"list"`
			}{
				List: []Template{
					{
						Type:  "datasource",
						Query: "prometheus",
					},
					{
						Type: "interval",
						Name: "interval",
						Options: []RawTemplateValue{
							map[string]interface{}{
								"value": "1h",
							},
						},
					},
					{
						Type:    "interval",
						Name:    "sampling",
						Current: map[string]interface{}{"value": "$__auto_interval_sampling"},
					},
					{
						Type: "resolution",
						Name: "resolution",
						Options: []RawTemplateValue{
							map[string]interface{}{
								"value": "1h",
							},
							map[string]interface{}{
								"value": "1h",
							},
						},
					},
				},
			},
			Panels: []Panel{
				tc.panel,
			},
		}

		testMultiResultRule(t, linter, dashboard, tc.result)
	}
}

// A dashboard can query more than one datasource: one datasource variable per
// datasource, and each target names the variable it queries. The rule checks
// the targets that query Prometheus and skips the ones that query another
// datasource, such as Loki.
func TestTargetPromQLRuleMixedDatasourceDashboards(t *testing.T) {
	linter := NewTargetPromQLRule()

	dashboard := Dashboard{
		Title: "mixed datasource dashboard",
		Templating: struct {
			List []Template `json:"list"`
		}{List: []Template{
			{Type: "datasource", Name: "prometheus_datasource", Query: "prometheus"},
			{Name: "loki_datasource", Type: "datasource", Query: "loki"},
		}},
		Panels: []Panel{
			{
				Title: "metrics",
				Type:  "timeseries",
				Targets: []Target{
					{Expr: `up{job=~"$job", instance=~"$instance"}`, Datasource: "${prometheus_datasource}"},
				},
			},
			{
				Title: "logs",
				Type:  "timeseries",
				Targets: []Target{
					{Expr: `{app="foo"} | json`, Datasource: "${loki_datasource}"},
				},
			},
		},
	}

	rs := ResultSet{}
	linter.Lint(dashboard, &rs)
	require.Len(t, rs.results, 2)
	for _, rc := range rs.results {
		for _, r := range rc.Result.Results {
			// A skipped target produces Quiet, a valid PromQL target
			// produces Success. Both mean the rule did not report the
			// target.
			if r.Severity != Quiet {
				require.Equal(t, ResultSuccess.Severity, r.Severity, "panel %s", rc.Panel.Title)
			}
		}
	}
}
