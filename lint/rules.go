package lint

type Rule interface {
	Description() string
	Name() string
	Lint(Dashboard, *ResultSet)
}

type DashboardRuleFunc struct {
	name, description string
	fn                func(Dashboard) DashboardRuleResults
}

func NewDashboardRuleFunc(name, description string, fn func(Dashboard) DashboardRuleResults) Rule {
	return &DashboardRuleFunc{name, description, fn}
}

func (f DashboardRuleFunc) Name() string        { return f.name }
func (f DashboardRuleFunc) Description() string { return f.description }
func (f DashboardRuleFunc) Lint(d Dashboard, s *ResultSet) {
	dashboardResults := f.fn(d).Results
	if len(dashboardResults) == 0 {
		dashboardResults = []DashboardResult{{
			Result: ResultSuccess,
		}}
	}
	rr := make([]FixableResult, len(dashboardResults))
	for i, r := range dashboardResults {
		r := r // capture loop variable
		var fix func(*Dashboard)
		if r.Fix != nil {
			fix = func(dashboard *Dashboard) {
				r.Fix(dashboard)
			}
		}
		rr[i] = FixableResult{
			Result: Result{
				Severity: r.Severity,
				Message:  r.Message,
			},
			Fix: fix,
		}
	}

	s.AddResult(ResultContext{
		Result:    RuleResults{rr},
		Rule:      f,
		Dashboard: &d,
	})
}

type PanelRuleFunc struct {
	name, description string
	fn                func(Dashboard, Panel) PanelRuleResults
}

func NewPanelRuleFunc(name, description string, fn func(Dashboard, Panel) PanelRuleResults) Rule {
	return &PanelRuleFunc{name, description, fn}
}

func (f PanelRuleFunc) Name() string        { return f.name }
func (f PanelRuleFunc) Description() string { return f.description }
func (f PanelRuleFunc) Lint(d Dashboard, s *ResultSet) {
	for pi, p := range d.GetPanels() {
		p := p   // capture loop variable
		pi := pi // capture loop variable
		var rr []FixableResult

		panelResults := f.fn(d, p).Results
		if len(panelResults) == 0 {
			panelResults = []PanelResult{{
				Result: ResultSuccess,
			}}
		}

		for _, r := range panelResults {
			var fix func(*Dashboard)
			if r.Fix != nil {
				fix = fixPanel(pi, r)
			}
			rr = append(rr, FixableResult{
				Result: Result{
					Severity: r.Severity,
					Message:  r.Message,
				},
				Fix: fix,
			})
		}

		s.AddResult(ResultContext{
			Result:    RuleResults{rr},
			Rule:      f,
			Dashboard: &d,
			Panel:     &p,
		})
	}
}

func fixPanel(pi int, r PanelResult) func(dashboard *Dashboard) {
	return func(dashboard *Dashboard) {
		p := dashboard.PanelAt(pi)
		if p == nil {
			return
		}
		cp := *p
		r.Fix(*dashboard, &cp)
		*p = cp
	}
}

type TargetRuleFunc struct {
	name, description string
	fn                func(Dashboard, Panel, Target) TargetRuleResults
}

func NewTargetRuleFunc(name, description string, fn func(Dashboard, Panel, Target) TargetRuleResults) Rule {
	return &TargetRuleFunc{name, description, fn}
}

func (f TargetRuleFunc) Name() string        { return f.name }
func (f TargetRuleFunc) Description() string { return f.description }
func (f TargetRuleFunc) Lint(d Dashboard, s *ResultSet) {
	for pi, p := range d.GetPanels() {
		p := p   // capture loop variable
		pi := pi // capture loop variable
		for ti, t := range p.Targets {
			t := t   // capture loop variable
			ti := ti // capture loop variable
			var rr []FixableResult

			targetResults := f.fn(d, p, t).Results
			if len(targetResults) == 0 {
				targetResults = []TargetResult{{
					Result: ResultSuccess,
				}}
			}

			for _, r := range targetResults {
				var fix func(*Dashboard)
				if r.Fix != nil {
					fix = fixTarget(pi, ti, r)
				}
				rr = append(rr, FixableResult{
					Result: Result{
						Severity: r.Severity,
						Message:  r.Message,
					},
					Fix: fix,
				})
			}
			s.AddResult(ResultContext{
				Result:    RuleResults{rr},
				Rule:      f,
				Dashboard: &d,
				Panel:     &p,
				Target:    &t,
			})
		}
	}
}

func fixTarget(pi int, ti int, r TargetResult) func(dashboard *Dashboard) {
	return func(dashboard *Dashboard) {
		p := dashboard.PanelAt(pi)
		if p == nil || ti >= len(p.Targets) {
			return
		}
		r.Fix(*dashboard, *p, &p.Targets[ti])
	}
}

// PanelAt returns a pointer to the panel at the given flattened index in the
// dashboard tree, matching the order produced by GetPanels (rows first, then
// top-level panels, depth-first). It is used by the autofix machinery to
// write fixes back to the correct location, including panels nested in rows
// or collapsible sections.
func (d *Dashboard) PanelAt(idx int) *Panel {
	i := 0
	var walk func(panels []Panel) *Panel
	walk = func(panels []Panel) *Panel {
		for j := range panels {
			if i == idx {
				return &panels[j]
			}
			i++
			if p := walk(panels[j].Panels); p != nil {
				return p
			}
		}
		return nil
	}
	for _, row := range d.Rows {
		if p := walk(row.Panels); p != nil {
			return p
		}
	}
	return walk(d.Panels)
}

// RuleSet contains a list of linting rules.
type RuleSet struct {
	rules []Rule
}

func NewRuleSet() RuleSet {
	return RuleSet{
		rules: []Rule{
			NewTemplateDatasourceRule(),
			NewTemplateJobRule(),
			NewTemplateInstanceRule(),
			NewTemplateLabelPromQLRule(),
			NewTemplateOnTimeRangeReloadRule(),
			NewPanelDatasourceRule(),
			NewPanelTitleDescriptionRule(),
			NewPanelUnitsRule(),
			NewPanelNoTargetsRule(),
			NewTargetLogQLRule(),
			NewTargetLogQLAutoRule(),
			NewTargetPromQLRule(),
			NewTargetRateIntervalRule(),
			NewTargetJobRule(),
			NewTargetInstanceRule(),
			NewTargetCounterAggRule(),
			NewUneditableRule(),
		},
	}
}

func (s *RuleSet) Rules() []Rule {
	return s.rules
}

func (s *RuleSet) Add(r Rule) {
	s.rules = append(s.rules, r)
}

func (s *RuleSet) Lint(dashboards []Dashboard) (*ResultSet, error) {
	resSet := &ResultSet{}
	for _, d := range dashboards {
		for _, r := range s.rules {
			r.Lint(d, resSet)
		}
	}
	return resSet, nil
}
