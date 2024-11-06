package lint

import (
	"testing"
)

func TestTargetLogQLRule(t *testing.T) {
	linter := NewTargetLogQLRule()

	for _, tc := range []struct {
		result Result
		panel  Panel
	}{
		// Don't fail non-Loki panels.
		{
			result: ResultSuccess,
			panel: Panel{
				Title:      "panel",
				Datasource: "prometheus",
				Targets: []Target{
					{
						Expr: `sum(rate(foo[5m]))`,
					},
				},
			},
		},
		// Valid LogQL query
		{
			result: ResultSuccess,
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate({job="mysql"}[5m]))`,
					},
				},
			},
		},
		// Invalid LogQL query
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'dashboard', panel 'panel', target idx '0' invalid LogQL query 'sum(rate({job="mysql"[5m]))': parse error at line 0, col 22: syntax error: unexpected RANGE, expecting } or ,`,
			},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate({job="mysql"[5m]))`,
					},
				},
			},
		},
		// Valid LogQL query with $__auto
		{
			result: ResultSuccess,
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate({job="mysql"}[$__auto]))`,
					},
				},
			},
		},
		// Valid complex LogQL query
		{
			result: ResultSuccess,
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum by (host) (rate({job="mysql"} |= "error" != "timeout" | json | duration > 10s [5m]))`,
					},
				},
			},
		},
		// Invalid complex LogQL query
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'dashboard', panel 'panel', target idx '0' invalid LogQL query 'sum by (host) (rate({job="mysql"} |= "error" != "timeout" | json | duration > 10s [5m])))': parse error at line 1, col 89: syntax error: unexpected )`,
			},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum by (host) (rate({job="mysql"} |= "error" != "timeout" | json | duration > 10s [5m])))`,
					},
				},
			},
		},
		// LogQL query with line_format
		{
			result: ResultSuccess,
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `{job="mysql"} | json | line_format "{{.timestamp}} {{.message}}"`,
					},
				},
			},
		},
		// LogQL query with unwrap
		{
			result: ResultSuccess,
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate({job="mysql"} | unwrap duration [5m]))`,
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
						Query: "loki",
					},
				},
			},
			Panels: []Panel{tc.panel},
		}
		testRule(t, linter, dashboard, tc.result)
	}
}
