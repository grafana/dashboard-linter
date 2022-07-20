package lint

import (
	"testing"
)

func TestTargetHistogramRule(t *testing.T) {
	linter := NewTargetHistogramRule()

	for _, tc := range []struct {
		result Result
		panel  Panel
	}{
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel '', target idx '0' histogram metric 'something_bucket' is not calculated in a histogram function",
			},
			panel: Panel{
				Targets: []Target{
					{
						Expr: `something_bucket`,
					},
				},
			},
		},
		// Is this one valid? Would someone ever do this, or must there always be a rate?
		{
			result: ResultSuccess,
			panel: Panel{
				Targets: []Target{
					{
						Expr: `histogram_quantile(0.9, something_bucket)`,
					},
				},
			},
		},
		{
			result: ResultSuccess,
			panel: Panel{
				Targets: []Target{
					{
						Expr: `histogram_quantile(0.9, rate(something_bucket[$__rate_interval]))`,
					},
				},
			},
		},
		{
			result: ResultSuccess,
			panel: Panel{
				Targets: []Target{
					{
						Expr: `histogram_quantile(0.9, sum by (le) (rate(something_bucket[$__rate_interval])))`,
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
