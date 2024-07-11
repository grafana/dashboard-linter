package lint

import (
	"testing"
)

func TestTargetRateIntervalRule(t *testing.T) {
	linter := NewTargetRateIntervalRule()

	for _, tc := range []struct {
		result Result
		panel  Panel
	}{
		// Don't fail non-prometheus panels.
		{
			result: ResultSuccess,
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
			result: ResultSuccess,
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate(foo{job=~"$job",instance=~"$instance"}[$__rate_interval]))`,
					},
				},
			},
		},
		// This is what a valid panel looks like.
		{
			result: ResultSuccess,
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate(foo{job=~"$job",instance=~"$instance"}[$__rate_interval]))/sum(rate(bar{job=~"$job",instance=~"$instance"}[$__rate_interval]))`,
					},
				},
			},
		},
		// Invalid query
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'sum(rate(foo{job=~"$job",instance=~"$instance"}[5m]))': should use $__rate_interval`,
			},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate(foo{job=~"$job",instance=~"$instance"}[5m]))`,
					},
				},
			},
		},
		// Timeseries support
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'sum(rate(foo{job=~"$job",instance=~"$instance"}[5m]))': should use $__rate_interval`,
			},
			panel: Panel{
				Title: "panel",
				Type:  "timeseries",
				Targets: []Target{
					{
						Expr: `sum(rate(foo{job=~"$job",instance=~"$instance"}[5m]))`,
					},
				},
			},
		},
		// Non-rate functions should not make the linter fail
		{
			result: ResultSuccess,
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(increase(foo{job=~"$job",instance=~"$instance"}[$__range]))`,
					},
				},
			},
		},
		// irate should be checked too
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'sum(irate(foo{job=~"$job",instance=~"$instance"}[$__interval]))': should use $__rate_interval`,
			},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(irate(foo{job=~"$job",instance=~"$instance"}[$__interval]))`,
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
				},
			},
			Panels: []Panel{tc.panel},
		}

		testRule(t, linter, dashboard, tc.result)
	}
}
