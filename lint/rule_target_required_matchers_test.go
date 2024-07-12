package lint

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/pkg/labels"
	"github.com/stretchr/testify/require"
)

func TestTargetRequiredMatcherRule(t *testing.T) {
	linter := NewTargetRequiredMatchersRule(&TargetRequiredMatchersRuleSettings{
		Matchers: config.Matchers{
			{
				Name:  "instance",
				Type:  labels.MatchRegexp,
				Value: "$instance",
			},
		},
	})

	for _, tc := range []struct {
		name   string
		result Result
		target Target
		fixed  *Target
	}{
		// Happy path
		{
			name:   "OK",
			result: ResultSuccess,
			target: Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s=~"$%s"}[5m]))`, "instance", "instance"),
			},
		},
		// Also happy when the promql is invalid
		{
			name:   "OK-invalid-promql",
			result: ResultSuccess,
			target: Target{
				Expr: `foo(bar.baz))`,
			},
		},
		// Missing matcher
		{
			name: "autofix-missing-matcher",
			result: Result{
				Severity: Fixed,
				Message:  fmt.Sprintf("Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'sum(rate(foo[5m]))': %s selector not found", "instance"),
			},
			target: Target{
				Expr: `sum(rate(foo[5m]))`,
			},
			fixed: &Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s=~"$%s"}[5m]))`, "instance", "instance"),
			},
		},
		// Not a regex matcher
		{
			name: "autofix-not-regex-matcher",
			result: Result{
				Severity: Fixed,
				Message:  fmt.Sprintf("Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'sum(rate(foo{%s=\"$%s\"}[5m]))': %s selector is =, not =~", "instance", "instance", "instance"),
			},
			target: Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s="$%s"}[5m]))`, "instance", "instance"),
			},
			fixed: &Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s=~"$%s"}[5m]))`, "instance", "instance"),
			},
		},
		// Wrong template variable
		{
			name: "autofix-wrong-template-variable",
			result: Result{
				Severity: Fixed,
				Message:  fmt.Sprintf("Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'sum(rate(foo{%s=~\"$foo\"}[5m]))': %s selector is $foo, not $%s", "instance", "instance", "instance"),
			},
			target: Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s=~"$foo"}[5m]))`, "instance"),
			},
			fixed: &Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s=~"$%s"}[5m]))`, "instance", "instance"),
			},
		},
		// Using Grafana global-variable
		{
			name: "autofix-reverse-expanded-variables",
			result: Result{
				Severity: Fixed,
				Message:  fmt.Sprintf("Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'sum(rate(foo[$__rate_interval]))': %s selector not found", "instance"),
			},
			target: Target{
				Expr: `sum(rate(foo[$__rate_interval]))`,
			},
			fixed: &Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s=~"$%s"}[$__rate_interval]))`, "instance", "instance"),
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
		t.Run(tc.name, func(t *testing.T) {
			autofix := tc.fixed != nil
			testRuleWithAutofix(t, linter, &dashboard, []Result{tc.result}, autofix)
			if autofix {
				fixedDashboard := Dashboard{
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
							Targets: []Target{*tc.fixed},
						},
					},
				}
				expected, _ := json.Marshal(fixedDashboard)
				actual, _ := json.Marshal(dashboard)
				require.Equal(t, string(expected), string(actual))
			}
		})
	}
}
