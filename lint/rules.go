package lint

type Rule interface {
	Description() string
	Name() string
}

type DashboardRule interface {
	Rule
	LintDashboard(Dashboard) Result
}

type DashboardRuleFunc struct {
	name, description string
	fn                func(Dashboard) Result
}

func (f DashboardRuleFunc) Name() string        { return f.name }
func (f DashboardRuleFunc) Description() string { return f.description }
func (f DashboardRuleFunc) LintDashboard(d Dashboard) Result {
	return f.fn(d)
}

type PanelRule interface {
	Rule
	LintPanel(Dashboard, Panel) Result
}

type PanelRuleFunc struct {
	name, description string
	fn                func(Dashboard, Panel) Result
}

func (f PanelRuleFunc) Name() string        { return f.name }
func (f PanelRuleFunc) Description() string { return f.description }
func (f PanelRuleFunc) LintPanel(d Dashboard, p Panel) Result {
	return f.fn(d, p)
}

type TargetRule interface {
	Rule
	LintTarget(Dashboard, Panel, Target) Result
}

// RuleSet contains a list of linting rules.
type RuleSet struct {
	dashboardRules []DashboardRule
	panelRules     []PanelRule
	targetRules    []TargetRule
}

func NewRuleSet() RuleSet {
	return RuleSet{
		dashboardRules: []DashboardRule{
			NewTemplateDatasourceRule(),
			NewTemplateJobRule(),
		},
		panelRules: []PanelRule{
			NewPanelDatasourceRule(),
			NewPanelPromQLRule(),
			NewPanelRateIntervalRule(),
		},
	}
}

func (s *RuleSet) Rules() []Rule {
	var result []Rule
	for i := range s.dashboardRules {
		result = append(result, s.dashboardRules[i])
	}
	for i := range s.panelRules {
		result = append(result, s.panelRules[i])
	}
	for i := range s.targetRules {
		result = append(result, s.targetRules[i])
	}
	return result
}

func (s *RuleSet) Lint(dashboards []Dashboard) (*ResultSet, error) {
	resSet := &ResultSet{}

	// Dashboards
	for _, dashboard := range dashboards {
		for _, dr := range s.dashboardRules {
			resSet.AddResult(ResultContext{
				Result:    dr.LintDashboard(dashboard),
				Rule:      dr,
				Dashboard: &dashboard,
			})
		}
		panels := dashboard.GetPanels()
		for p := range panels {
			for _, pr := range s.panelRules {
				resSet.AddResult(ResultContext{
					Result:    pr.LintPanel(dashboard, panels[p]),
					Rule:      pr,
					Dashboard: &dashboard,
					Panel:     &panels[p],
				})
			}
			targets := panels[p].Targets
			for t := range targets {
				for _, tr := range s.targetRules {
					resSet.AddResult(ResultContext{
						Result:    tr.LintTarget(dashboard, panels[p], targets[t]),
						Rule:      tr,
						Dashboard: &dashboard,
						Panel:     &panels[p],
						Target:    &targets[t],
					})
				}
			}
		}
	}
	return resSet, nil
}
