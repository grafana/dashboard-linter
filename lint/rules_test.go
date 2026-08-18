package lint_test

import (
	"os"
	"testing"

	"github.com/grafana/dashboard-linter/lint"
	"github.com/stretchr/testify/assert"
)

func TestCustomRules(t *testing.T) {
	sampleDashboard, err := os.ReadFile("testdata/dashboard.json")
	assert.NoError(t, err)

	for _, tc := range []struct {
		desc string
		rule lint.Rule
	}{
		{
			desc: "Should allow addition of dashboard rule",
			rule: lint.NewDashboardRuleFunc(
				"test-dashboard-rule", "Test dashboard rule",
				func(lint.Dashboard) lint.DashboardRuleResults {
					return lint.DashboardRuleResults{Results: []lint.DashboardResult{{
						Result: lint.Result{Severity: lint.Error, Message: "Error found"},
					}}}
				},
			),
		},
		{
			desc: "Should allow addition of panel rule",
			rule: lint.NewPanelRuleFunc(
				"test-panel-rule", "Test panel rule",
				func(d lint.Dashboard, p lint.Panel) lint.PanelRuleResults {
					return lint.PanelRuleResults{Results: []lint.PanelResult{{
						Result: lint.Result{Severity: lint.Error, Message: "Error found"},
					}}}
				},
			),
		},
		{
			desc: "Should allow addition of target rule",
			rule: lint.NewTargetRuleFunc(
				"test-target-rule", "Test target rule",
				func(lint.Dashboard, lint.Panel, lint.Target) lint.TargetRuleResults {
					return lint.TargetRuleResults{Results: []lint.TargetResult{{
						Result: lint.Result{Severity: lint.Error, Message: "Error found"},
					}}}
				},
			),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			rules := lint.RuleSet{}
			rules.Add(tc.rule)

			dashboard, err := lint.NewDashboard(sampleDashboard)
			assert.NoError(t, err)

			results, err := rules.Lint([]lint.Dashboard{dashboard})
			assert.NoError(t, err)

			// Validate the error was added
			assert.GreaterOrEqual(t, len(results.ByRule()[tc.rule.Name()]), 1)
			r := results.ByRule()[tc.rule.Name()][0].Result
			assert.Equal(t, lint.Result{Severity: lint.Error, Message: "Error found"}, r.Results[0].Result)
		})
	}
}

func TestFixableRules(t *testing.T) {
	sampleDashboard, err := os.ReadFile("testdata/dashboard.json")
	assert.NoError(t, err)

	rule := lint.NewDashboardRuleFunc(
		"test-fixable-rule", "Test fixable rule",
		func(d lint.Dashboard) lint.DashboardRuleResults {
			rr := lint.DashboardRuleResults{}
			rr.AddFixableError(d, "fixing first issue", func(d *lint.Dashboard) {
				d.Title += " fixed-once"
			})
			rr.AddFixableError(d, "fixing second issue", func(d *lint.Dashboard) {
				d.Title += " fixed-twice"
			})
			return rr
		},
	)

	rules := lint.RuleSet{}
	rules.Add(rule)

	dashboard, err := lint.NewDashboard(sampleDashboard)
	assert.NoError(t, err)

	results, err := rules.Lint([]lint.Dashboard{dashboard})
	assert.NoError(t, err)

	results.AutoFix(&dashboard)

	assert.Equal(t, "Sample dashboard fixed-once fixed-twice", dashboard.Title)
}

func TestFixableRulesNestedPanels(t *testing.T) {
	dashboardJSON := `{
		"title": "nested",
		"rows": [{"panels": [
			{"id": 1, "title": "row-panel", "type": "timeseries",
			 "targets": [{"expr": "up", "refId": "A"}]}
		]}],
		"panels": [{"id": 2, "title": "top-panel", "type": "timeseries", "panels": [
			{"id": 3, "title": "nested-panel", "type": "timeseries",
			 "targets": [{"expr": "up", "refId": "A"}]}
		]}]
	}`

	panelRule := lint.NewPanelRuleFunc(
		"test-panel-fix-rule", "Test panel fix rule",
		func(d lint.Dashboard, p lint.Panel) lint.PanelRuleResults {
			return lint.PanelRuleResults{Results: []lint.PanelResult{{
				Result: lint.Result{Severity: lint.Error, Message: "fixing panel"},
				Fix: func(d lint.Dashboard, p *lint.Panel) {
					p.Title += "-fixed"
				},
			}}}
		},
	)
	targetRule := lint.NewTargetRuleFunc(
		"test-target-fix-rule", "Test target fix rule",
		func(d lint.Dashboard, p lint.Panel, t lint.Target) lint.TargetRuleResults {
			return lint.TargetRuleResults{Results: []lint.TargetResult{{
				Result: lint.Result{Severity: lint.Error, Message: "fixing target"},
				Fix: func(d lint.Dashboard, p lint.Panel, t *lint.Target) {
					t.Expr = "up{job=\"$job\"}"
				},
			}}}
		},
	)

	rules := lint.RuleSet{}
	rules.Add(panelRule)
	rules.Add(targetRule)

	dashboard, err := lint.NewDashboard([]byte(dashboardJSON))
	assert.NoError(t, err)
	panels := dashboard.GetPanels()
	assert.Len(t, panels, 3)

	results, err := rules.Lint([]lint.Dashboard{dashboard})
	assert.NoError(t, err)
	results.AutoFix(&dashboard)

	// Правки должны попасть в исходное дерево: в строку, в топ-уровень
	// и во вложенную панель, а не только в первый слой dashboard.Panels.
	assert.Equal(t, "row-panel-fixed", dashboard.Rows[0].Panels[0].Title)
	assert.Equal(t, "top-panel-fixed", dashboard.Panels[0].Title)
	assert.Equal(t, "nested-panel-fixed", dashboard.Panels[0].Panels[0].Title)
	assert.Equal(t, "up{job=\"$job\"}", dashboard.Rows[0].Panels[0].Targets[0].Expr)
	assert.Equal(t, "up{job=\"$job\"}", dashboard.Panels[0].Panels[0].Targets[0].Expr)
}

func TestPanelAt(t *testing.T) {
	dashboardJSON := `{
		"title": "nested",
		"rows": [{"panels": [
			{"id": 1, "title": "row-panel", "type": "timeseries",
			 "panels": [{"id": 11, "title": "row-nested", "type": "timeseries"}]}
		]}],
		"panels": [
			{"id": 2, "title": "top-panel", "type": "timeseries",
			 "panels": [{"id": 22, "title": "top-nested", "type": "timeseries"}]},
			{"id": 3, "title": "last-panel", "type": "timeseries"}
		]
	}`

	dashboard, err := lint.NewDashboard([]byte(dashboardJSON))
	assert.NoError(t, err)

	// Индексы должны соответствовать порядку GetPanels: строки сначала,
	// потом топ-уровневые панели, в глубину.
	for i, want := range []string{"row-panel", "row-nested", "top-panel", "top-nested", "last-panel"} {
		assert.Equal(t, want, dashboard.PanelAt(i).Title, "index %d", i)
	}
}
