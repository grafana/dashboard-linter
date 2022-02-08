package lint

import (
	"testing"

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

func TestCustomRules(t *testing.T) {
	for _, tc := range []struct {
		desc string
		rule Rule
	}{
		{
			desc: "Should allow addition of dashboard rule",
			rule: DashboardRuleFunc{
				fn: func(Dashboard) Result {
					return Result{Severity: Error, Message: "Error found"}
				},
			},
		},
		{
			desc: "Should allow addition of panel rule",
			rule: PanelRuleFunc{
				fn: func(Dashboard, Panel) Result {
					return Result{Severity: Error, Message: "Error found"}
				},
			},
		},
		{
			desc: "Should allow addition of target rule",
			rule: TargetRuleFunc{
				fn: func(Dashboard, Panel, Target) Result {
					return Result{Severity: Error, Message: "Error found"}
				},
			},
		},
	} {
		rules := RuleSet{
			rules: []Rule{tc.rule},
		}

		dashboard, err := NewDashboard([]byte(dashboard))
		assert.NoError(t, err, tc.desc)

		results, err := rules.Lint([]Dashboard{dashboard})
		assert.NoError(t, err, tc.desc)

		// Validate the error was added
		assert.Len(t, results.results, 1)
		assert.Equal(t, results.results[0].Result, Result{Severity: Error, Message: "Error found"})
	}
}
