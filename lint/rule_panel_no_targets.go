package lint

import "fmt"

func NewPanelNoTargetsRule() *PanelRuleFunc {

	return &PanelRuleFunc{
		name:        "panel-no-targets-rule",
		description: "Checks that each panel has at least one target.",
		fn: func(d Dashboard, p Panel) Result {
			switch p.Type {
			case "stat", "singlestat", "graph", "table", "timeseries", "gauge":
				if p.Targets != nil {
					return ResultSuccess
				}

				return Result{
					Severity: Error,
					Message:  fmt.Sprintf("Dashboard '%s', panel '%s' has no targets", d.Title, p.Title),
				}
			}
			return ResultSuccess
		},
	}
}
