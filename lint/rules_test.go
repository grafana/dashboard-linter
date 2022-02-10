package lint_test

import (
	"io/ioutil"
	"testing"

	"github.com/grafana/dashboard-linter/lint"
	"github.com/stretchr/testify/assert"
)

func TestCustomRules(t *testing.T) {
	sampleDashboard, err := ioutil.ReadFile("testdata/dashboard.json")
	assert.NoError(t, err)

	for _, tc := range []struct {
		desc string
		rule lint.Rule
	}{
		{
			desc: "Should allow addition of dashboard rule",
			rule: lint.NewDashboardRuleFunc(
				"test-dashboard-rule", "Test dashboard rule",
				func(lint.Dashboard) lint.Result {
					return lint.Result{Severity: lint.Error, Message: "Error found"}
				},
			),
		},
		{
			desc: "Should allow addition of panel rule",
			rule: lint.NewPanelRuleFunc(
				"test-panel-rule", "Test panel rule",
				func(d lint.Dashboard, p lint.Panel) lint.Result {
					return lint.Result{Severity: lint.Error, Message: "Error found"}
				},
			),
		},
		{
			desc: "Should allow addition of target rule",
			rule: lint.NewTargetRuleFunc(
				"test-target-rule", "Test target rule",
				func(lint.Dashboard, lint.Panel, lint.Target) lint.Result {
					return lint.Result{Severity: lint.Error, Message: "Error found"}
				},
			),
		},
	} {
		rules := lint.RuleSet{}
		rules.Add(tc.rule)

		dashboard, err := lint.NewDashboard(sampleDashboard)
		assert.NoError(t, err, tc.desc)

		results, err := rules.Lint([]lint.Dashboard{dashboard})
		assert.NoError(t, err, tc.desc)

		// Validate the error was added
		assert.GreaterOrEqual(t, len(results.ByRule()[tc.rule.Name()]), 1)
		assert.Equal(t, results.ByRule()[tc.rule.Name()][0].Result, lint.Result{Severity: lint.Error, Message: "Error found"})
	}
}
