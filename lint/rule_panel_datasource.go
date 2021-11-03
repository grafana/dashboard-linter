package lint

import (
	"fmt"
)

func NewPanelDatasourceRule() *PanelRuleFunc {
	return &PanelRuleFunc{
		name:        "panel-datasource-rule",
		description: "panel-datasource-rule Checks that each panel uses the templated datasource.",
		fn: func(d Dashboard, p Panel) Result {

			switch p.Type {
			case "singlestat", "graph", "table":
				if p.Datasource != "$datasource" && p.Datasource != "${datasource}" {
					return Result{
						Severity: Error,
						Message:  fmt.Sprintf("Dashboard '%s', panel '%s' does not use templates datasource, uses '%s'", d.Title, p.Title, p.Datasource),
					}
				}
			}

			return Result{
				Severity: Success,
				Message:  "OK",
			}
		},
	}
}
