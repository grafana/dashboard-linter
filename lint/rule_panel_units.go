package lint

import "fmt"

func NewPanelUnitsRule() *PanelRuleFunc {
	return &PanelRuleFunc{
		name:        "panel-units-rule",
		description: "Checks that each panel uses has valid units defined.",
		fn: func(d Dashboard, p Panel) Result {
			switch p.Type {
			case "stat", "singlestat", "graph", "table", "timeseries":
				if len(p.FieldConfig.Defaults.Unit) > 0 {
					switch p.FieldConfig.Defaults.Unit {
					case "short":
						return ResultSuccess
					}
				}
				return Result{
					Severity: Error,
					Message:  fmt.Sprintf("Dashboard '%s', panel '%s' has no or invalid units defined: '%s'", d.Title, p.Title, p.FieldConfig.Defaults.Unit),
				}
			}
			return ResultSuccess
		},
	}
}
