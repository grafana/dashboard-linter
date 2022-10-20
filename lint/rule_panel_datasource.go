package lint

import (
	"fmt"
	"regexp"
)

func NewPanelDatasourceRule() *PanelRuleFunc {
	return &PanelRuleFunc{
		name:        "panel-datasource-rule",
		description: "Checks that each panel uses the templated datasource.",
		fn: func(d Dashboard, p Panel) Result {
			switch p.Type {
			case "singlestat", "graph", "table", "timeseries":
				match, _ := regexp.MatchString("^\\${?.+}?", string(p.Datasource))
				if !match {
					return Result{
						Severity: Error,
						Message:  fmt.Sprintf("Dashboard '%s', panel '%s' does not use templates datasource, uses '%s'", d.Title, p.Title, p.Datasource),
					}
				}
			}

			return ResultSuccess
		},
	}
}
