package lint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPanelPromQLRule(t *testing.T) {
	linter := NewPanelPromQLRule()
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
	}

	for _, tc := range []struct {
		result Result
		panel  Panel
	}{
		// Don't fail non-prometheus panels.
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			panel: Panel{
				Title:      "panel",
				Datasource: "foo",
			},
		},
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
						Expr: `sum(rate(foo[5m]))`,
					},
				},
			},
		},
		// Invalid query
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel' invalid PromQL query 'foo(bar.baz)': 1:8: parse error: unexpected character: '.'",
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
		// Timeseries support
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel' invalid PromQL query 'foo(bar.baz)': 1:8: parse error: unexpected character: '.'",
			},
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
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
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
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
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
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
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
	} {
		require.Equal(t, tc.result, linter.LintPanel(dashboard, tc.panel))
	}
}
