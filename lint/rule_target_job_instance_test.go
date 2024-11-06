package lint

import (
	"fmt"
	"testing"
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
		result Result
		target Target
	}{
		// Happy path
		{
			result: ResultSuccess,
			target: Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s=~"$%s"}[5m]))`, matcher, matcher),
			},
		},
		// Happy path (multiple matchers where at least one matches)
		{
			result: ResultSuccess,
			target: Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s="integrations/bar", %s=~"$%s"}[5m]))`, matcher, matcher, matcher),
			},
		},
		{
			result: ResultSuccess,
			target: Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s=~"$%s", %s="integrations/bar"}[5m]))`, matcher, matcher, matcher),
			},
		},
		// Also happy when the promql is invalid
		{
			result: ResultSuccess,
			target: Target{
				Expr: `foo(bar.baz))`,
			},
		},
		// Missing matcher
		{
			result: Result{
				Severity: Error,
				Message:  fmt.Sprintf("Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'sum(rate(foo[5m]))': %s selector not found", matcher),
			},
			target: Target{
				Expr: `sum(rate(foo[5m]))`,
			},
		},
		// Not a regex matcher
		{
			result: Result{
				Severity: Error,
				Message:  fmt.Sprintf("Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'sum(rate(foo{%s=\"$%s\"}[5m]))': %s selector is =, not =~", matcher, matcher, matcher),
			},
			target: Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s="$%s"}[5m]))`, matcher, matcher),
			},
		},
		// Wrong template variable
		{
			result: Result{
				Severity: Error,
				Message:  fmt.Sprintf("Dashboard 'dashboard', panel 'panel', target idx '0' invalid PromQL query 'sum(rate(foo{%s=~\"$foo\"}[5m]))': %s selector is $foo, not $%s", matcher, matcher, matcher),
			},
			target: Target{
				Expr: fmt.Sprintf(`sum(rate(foo{%s=~"$foo"}[5m]))`, matcher),
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

func TestTargetJobInstanceRule(t *testing.T) {
	testTargetRequiredMatcherRule(t, "job")
	testTargetRequiredMatcherRule(t, "instance")
}
