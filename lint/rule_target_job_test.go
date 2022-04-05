package lint

import (
	"testing"
)

func TestTargetJobRule(t *testing.T) {
	linter := NewTargetJobRule()

	for _, tc := range []struct {
		result Result
		target Target
	}{
		// Happy path
		{
			result: ResultSuccess,
			target: Target{
				Expr: `sum(rate(foo{job=~"$job"}[5m]))`,
			},
		},
		// Also happy when the promql is invalid
		{
			result: ResultSuccess,
			target: Target{
				Expr: `foo(bar.baz))`,
			},
		},
		// Missing job matcher
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'sum(rate(foo[5m]))': job selector not found",
			},
			target: Target{
				Expr: `sum(rate(foo[5m]))`,
			},
		},
		// Not a regex matcher
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'sum(rate(foo{job=\"$job\"}[5m]))': job selector is =, not =~",
			},
			target: Target{
				Expr: `sum(rate(foo{job="$job"}[5m]))`,
			},
		},
		// Wrong template variable
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'sum(rate(foo{job=~\"$foo\"}[5m]))': job selector is $foo, not $job",
			},
			target: Target{
				Expr: `sum(rate(foo{job=~"$foo"}[5m]))`,
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
			Panels: []Panel{
				{
					Title:   "panel",
					Type:    "singlestat",
					Targets: []Target{tc.target},
				},
			},
		}

		testRule(t, linter, dashboard, tc.result)
	}
}
