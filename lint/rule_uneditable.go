package lint

func NewUneditableRule() *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "uneditable-dashboard",
		description: "Checks that the dashboard is not editable.",
		fn: func(d Dashboard) DashboardRuleResults {
			r := DashboardRuleResults{}
			if d.Editable {
				r.AddFixableError(d, "is editable, it should be set to 'editable: false'", FixUneditableRule)
			}
			return r
		},
	}
}

func FixUneditableRule(d *Dashboard) {
	d.Editable = false
}
