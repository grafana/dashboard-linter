package lint

type Rule interface {
	Description() string
	Name() string
	Lint(Dashboard, *ResultSet)
}

type DashboardRuleFunc struct {
	name, description string
	fn                func(Dashboard) Result
}

func NewDashboardRuleFunc(name, description string, fn func(Dashboard) Result) Rule {
	return &DashboardRuleFunc{name, description, fn}
}

func (f DashboardRuleFunc) Name() string        { return f.name }
func (f DashboardRuleFunc) Description() string { return f.description }
func (f DashboardRuleFunc) Lint(d Dashboard, s *ResultSet) {
	s.AddResult(ResultContext{
		Result:    f.fn(d),
		Rule:      f,
		Dashboard: &d,
	})
}

type PanelRuleFunc struct {
	name, description string
	fn                func(Dashboard, Panel) Result
}

func NewPanelRuleFunc(name, description string, fn func(Dashboard, Panel) Result) Rule {
	return &PanelRuleFunc{name, description, fn}
}

func (f PanelRuleFunc) Name() string        { return f.name }
func (f PanelRuleFunc) Description() string { return f.description }
func (f PanelRuleFunc) Lint(d Dashboard, s *ResultSet) {
	for _, p := range d.GetPanels() {
		p := p // capture loop variable
		s.AddResult(ResultContext{
			Result:    f.fn(d, p),
			Rule:      f,
			Dashboard: &d,
			Panel:     &p,
		})
	}
}

type TargetRuleFunc struct {
	name, description string
	fn                func(Dashboard, Panel, Target) Result
}

func NewTargetRuleFunc(name, description string, fn func(Dashboard, Panel, Target) Result) Rule {
	return &TargetRuleFunc{name, description, fn}
}

func (f TargetRuleFunc) Name() string        { return f.name }
func (f TargetRuleFunc) Description() string { return f.description }
func (f TargetRuleFunc) Lint(d Dashboard, s *ResultSet) {
	for _, p := range d.GetPanels() {
		p := p // capture loop variable
		for _, t := range p.Targets {
			t := t // capture loop variable
			s.AddResult(ResultContext{
				Result:    f.fn(d, p, t),
				Rule:      f,
				Dashboard: &d,
				Panel:     &p,
				Target:    &t,
			})
		}
	}
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
			NewPanelDatasourceRule(),
			NewTargetPromQLRule(),
			NewTargetRateIntervalRule(),
			NewTargetJobRule(),
			NewTargetInstanceRule(),
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
