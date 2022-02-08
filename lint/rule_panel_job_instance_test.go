package lint

import (
	"testing"
)

func TestPanelJobInstanceRule(t *testing.T) {
	linter := NewPanelJobInstanceRule()

	for _, tc := range []struct {
		result Result
		panel  Panel
	}{
		// This is what a valid panel looks like.
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
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
		// Invalid query should be fine.
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `foo(bar.baz)`,
					},
				},
			},
		},
		// Missing job matcher
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel' invalid PromQL query 'sum(rate(foo[5m]))': job selector not found",
			},
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
		// Missing instance matcher
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel' invalid PromQL query 'sum(rate(foo{job=~\"$job\"}[5m]))': instance selector not found",
			},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate(foo{job=~"$job"}[5m]))`,
					},
				},
			},
		},
		// Not a regex matcher
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel' invalid PromQL query 'sum(rate(foo{job=\"$job\",instance=\"$instance\"}[5m]))': job selector is =, not =~",
			},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate(foo{job="$job",instance="$instance"}[5m]))`,
					},
				},
			},
		},
		// Wrong template variable.
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel' invalid PromQL query 'sum(rate(foo{job=~\"$instance\",instance=~\"$job\"}[5m]))': job selector is $instance, not $job",
			},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate(foo{job=~"$instance",instance=~"$job"}[5m]))`,
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
