package lint_test

import (
	"testing"

	"github.com/grafana/dashboard-linter/lint"
	"github.com/stretchr/testify/assert"
)

const dashboard = `{
	"panels": [
		{
			"type": "timeseries",
			"title": "Timeseries",
			"targets": [
				{
					"expr": "up{job=\"$job\"}"
				}
			]
		}
	],
	"templating": {
		"list": [
			{
				"current": {
					"text": "default",
					"value": "default"
				},
				"hide": 0,
				"label": "Data Source",
				"name": "datasource",
				"options": [ ],
				"query": "prometheus",
				"refresh": 1,
				"regex": "",
				"type": "datasource"
			},
			{
				"name": "job",
				"label": "job",
				"datasource": "$datasource",
				"type": "query",
				"query": "query_result(up{})",
				"multi": true,
				"allValue": ".+"
			}
		]
	},
	"title": "Sample dashboard"
}`

type Rule struct {
	RuleName string
}

func (rule *Rule) Name() string {
	return rule.RuleName
}

func (rule *Rule) Description() string {
	return "Test rule"
}

type CustomDashboardRule struct {
	Rule
}

func (rule *CustomDashboardRule) LintDashboard(dashboard lint.Dashboard) lint.Result {
	return lint.Result{Severity: lint.Error, Message: "Error found"}
}

type CustomPanelRule struct {
	Rule
}

func (rule *CustomPanelRule) LintPanel(dashboard lint.Dashboard, panel lint.Panel) lint.Result {
	return lint.Result{Severity: lint.Error, Message: "Error found"}
}

type CustomTargetRule struct {
	Rule
}

func (rule *CustomTargetRule) LintTarget(dashboard lint.Dashboard, panel lint.Panel, target lint.Target) lint.Result {
	return lint.Result{Severity: lint.Error, Message: "Error found"}
}

func TestCustomRules(t *testing.T) {
	for _, tc := range []struct {
		desc string
		rule lint.Rule
	}{
		{
			desc: "Should allow addition of dashboard rule",
			rule: &CustomDashboardRule{Rule: Rule{RuleName: "test-dashboard-rule"}},
		},
		{
			desc: "Should allow addition of panel rule",
			rule: &CustomPanelRule{Rule: Rule{RuleName: "test-panel-rule"}},
		},
		{
			desc: "Should allow addition of target rule",
			rule: &CustomTargetRule{Rule: Rule{RuleName: "test-target-rule"}},
		},
	} {
		rules := lint.NewRuleSet()
		// Add custom rule
		rules.Add(tc.rule)
		// Lint
		dashboard, err := lint.NewDashboard([]byte(dashboard))
		assert.NoError(t, err, tc.desc)
		result, err := rules.Lint([]lint.Dashboard{dashboard})
		assert.NoError(t, err, tc.desc)
		// Validate the error was added
		results := result.ByRule()
		assert.Len(t, results[tc.rule.Name()], 1)
		for _, result := range results[tc.rule.Name()] {
			assert.Equal(t, result.Result, lint.Result{Severity: lint.Error, Message: "Error found"})
		}
	}
}
