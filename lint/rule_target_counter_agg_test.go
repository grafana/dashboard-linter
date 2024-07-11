package lint

import (
	"testing"
)

func TestTargetCounterAggRule(t *testing.T) {
	linter := NewTargetCounterAggRule()

	for _, tc := range []struct {
		result Result
		panel  Panel
	}{
		// Non aggregated counter fails
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel', target idx '0' counter metric 'something_total' is not aggregated with rate, irate, or increase",
			},
			panel: Panel{
				Title:      "panel",
				Datasource: "foo",
				Targets: []Target{
					{
						Expr: `something_total`,
					},
				},
			},
		},
		// Weird matrix selector without an aggregator
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel', target idx '0' counter metric 'something_total' is not aggregated with rate, irate, or increase",
			},
			panel: Panel{
				Title:      "panel",
				Datasource: "foo",
				Targets: []Target{
					{
						Expr: `something_total[$__rate_interval]`,
					},
				},
			},
		},
		// Single aggregated counter is good
		{
			result: ResultSuccess,
			panel: Panel{
				Title:      "panel",
				Datasource: "foo",
				Targets: []Target{
					{
						Expr: `increase(something_total[$__rate_interval])`,
					},
				},
			},
		},
		// Sanity check for multiple counters in one query, with the first one failing
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel', target idx '0' counter metric 'something_total' is not aggregated with rate, irate, or increase",
			},
			panel: Panel{
				Title:      "panel",
				Datasource: "foo",
				Targets: []Target{
					{
						Expr: `something_total / rate(somethingelse_total[$__rate_interval])`,
					},
				},
			},
		},
		// Sanity check for multiple counters in one query, with the second one failing
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel', target idx '0' counter metric 'somethingelse_total' is not aggregated with rate, irate, or increase",
			},
			panel: Panel{
				Title:      "panel",
				Datasource: "foo",
				Targets: []Target{
					{
						Expr: `rate(something_total[$__rate_interval]) / somethingelse_total`,
					},
				},
			},
		},
	} {
		dashboard := Dashboard{
			Title: "dashboard",
			Templating: struct {
				List []Template "json:\"list\""
			}{List: []Template{}},
			Panels: []Panel{tc.panel},
		}

		testRule(t, linter, dashboard, tc.result)
	}
}
