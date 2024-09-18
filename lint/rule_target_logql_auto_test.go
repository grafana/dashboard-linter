package lint

import (
	"testing"
)

// TestTargetLogQLAutoRule tests the NewTargetLogQLAutoRule function to ensure
// that it correctly identifies LogQL queries that should use $__auto for range vectors.
func TestTargetLogQLAutoRule(t *testing.T) {
	linter := NewTargetLogQLAutoRule()

	for _, tc := range []struct {
		result Result
		panel  Panel
	}{
		// Test case: Non-Loki panel should pass without errors.
		{
			result: ResultSuccess,
			panel: Panel{
				Title:      "panel",
				Datasource: "foo",
				Targets: []Target{
					{
						Expr: `sum(rate({job=~"$job",instance=~"$instance"}[5m]))`,
					},
				},
			},
		},
		// Test case: Valid LogQL query using $__auto.
		{
			result: ResultSuccess,
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate({job=~"$job",instance=~"$instance"} [$__auto]))`,
					},
				},
			},
		},
		// Test case: Valid LogQL query using $__auto in a complex expression.
		{
			result: ResultSuccess,
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate({job=~"$job",instance=~"$instance"} [$__auto]))/sum(rate({job=~"$job",instance=~"$instance"} [$__auto]))`,
					},
				},
			},
		},
		// Test case: Invalid LogQL query without $__auto.
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'dashboard', panel 'panel', target idx '0' LogQL query uses fixed duration: should use $__auto`,
			},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate({job=~"$job",instance=~"$instance"}[5m]))`,
					},
				},
			},
		},
		// Test case: Invalid LogQL query without $__auto in a timeseries panel.
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'dashboard', panel 'panel', target idx '0' LogQL query uses fixed duration: should use $__auto`,
			},
			panel: Panel{
				Title: "panel",
				Type:  "timeseries",
				Targets: []Target{
					{
						Expr: `sum(rate({job=~"$job",instance=~"$instance"}[5m]))`,
					},
				},
			},
		},
		// Test case: Valid LogQL query with count_over_time and $__auto.
		{
			result: ResultSuccess,
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `count_over_time({job="mysql"} [$__auto])`,
					},
				},
			},
		},
		// Test case: Invalid LogQL query with count_over_time without $__auto.
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'dashboard', panel 'panel', target idx '0' LogQL query uses fixed duration: should use $__auto`,
			},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `count_over_time({job="mysql"}[5m])`,
					},
				},
			},
		},
		// Test case: Valid LogQL query with bytes_rate and $__auto.
		{
			result: ResultSuccess,
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `bytes_rate({job="mysql"} [$__auto])`,
					},
				},
			},
		},
		// Test case: Invalid LogQL query with bytes_rate without $__auto.
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'dashboard', panel 'panel', target idx '0' LogQL query uses fixed duration: should use $__auto`,
			},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `bytes_rate({job="mysql"}[5m])`,
					},
				},
			},
		},
		// Test case: Valid LogQL query with bytes_over_time and $__auto.
		{
			result: ResultSuccess,
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `bytes_over_time({job="mysql"} [$__auto])`,
					},
				},
			},
		},
		// Test case: Invalid LogQL query with bytes_over_time without $__auto.
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'dashboard', panel 'panel', target idx '0' LogQL query uses fixed duration: should use $__auto`,
			},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `bytes_over_time({job="mysql"}[5m])`,
					},
				},
			},
		},
		// Test case: Valid LogQL query with sum_over_time and $__auto.
		{
			result: ResultSuccess,
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum_over_time({job="mysql"} |= "duration" | unwrap duration [$__auto])`,
					},
				},
			},
		},
		// Test case: Invalid LogQL query with sum_over_time without $__auto.
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'dashboard', panel 'panel', target idx '0' LogQL query uses fixed duration: should use $__auto`,
			},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum_over_time({job="mysql"} |= "duration" | unwrap duration[5m])`,
					},
				},
			},
		},
		// Test case: Valid LogQL query with avg_over_time and $__auto.
		{
			result: ResultSuccess,
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `avg_over_time({job="mysql"} |= "duration" | unwrap duration [$__auto])`,
					},
				},
			},
		},
		// Test case: Invalid LogQL query with avg_over_time without $__auto.
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'dashboard', panel 'panel', target idx '0' LogQL query uses fixed duration: should use $__auto`,
			},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `avg_over_time({job="mysql"} |= "duration" | unwrap duration[5m])`,
					},
				},
			},
		},
		// Add similar tests for other unwrapped range aggregations...
	} {
		dashboard := Dashboard{
			Title: "dashboard",
			Templating: struct {
				List []Template `json:"list"`
			}{
				List: []Template{
					{
						Type:  "datasource",
						Query: "loki",
					},
				},
			},
			Panels: []Panel{tc.panel},
		}
		testRule(t, linter, dashboard, tc.result)
	}
}
