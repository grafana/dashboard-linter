package lint

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func testTargetRequiredMatcherRule(t *testing.T, matcher string) {
	var linter *TargetRuleFunc

	switch matcher {
	case "job":
		linter = NewTargetJobRule()
	case "instance":
		linter = NewTargetInstanceRule()
	default:
		t.Errorf("No concrete target required matcher rule for '%s", matcher)
		return
	}

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
				Expr: fmt.Sprintf(`sum(rate(foo{%s=~"$%s"}[5m]))`, matcher, matcher),
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
				Message:  fmt.Sprintf("Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'sum(rate(foo[5m]))': %s selector not found", matcher),
			},
			target: Target{
				Expr: `sum(rate(foo[5m]))`,
			},
			fixed: &Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s=~"$%s"}[5m]))`, matcher, matcher),
			},
		},
		// Not a regex matcher
		{
			name: "autofix-not-regex-matcher",
			result: Result{
				Severity: Fixed,
				Message:  fmt.Sprintf("Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'sum(rate(foo{%s=\"$%s\"}[5m]))': %s selector is =, not =~", matcher, matcher, matcher),
			},
			target: Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s="$%s"}[5m]))`, matcher, matcher),
			},
			fixed: &Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s=~"$%s"}[5m]))`, matcher, matcher),
			},
		},
		// Wrong template variable
		{
			name: "autofix-wrong-template-variable",
			result: Result{
				Severity: Fixed,
				Message:  fmt.Sprintf("Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'sum(rate(foo{%s=~\"$foo\"}[5m]))': %s selector is $foo, not $%s", matcher, matcher, matcher),
			},
			target: Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s=~"$foo"}[5m]))`, matcher),
			},
			fixed: &Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s=~"$%s"}[5m]))`, matcher, matcher),
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

func TestTargetJobInstanceRule(t *testing.T) {
	testTargetRequiredMatcherRule(t, "job")
	testTargetRequiredMatcherRule(t, "instance")
}
