package lint

import (
	"fmt"
	"strings"
)

func NewPanelDatasourceRule() *PanelRuleFunc {
	return &PanelRuleFunc{
		name:        "panel-datasource-rule",
		description: "Checks that each panel uses the templated datasource.",
		fn: func(d Dashboard, p Panel) Result {
			switch p.Type {
			case "singlestat", "graph", "table", "timeseries":

				if expectedDatasources := checkTemplatedDatasourceUsed(d, p.Datasource); len(expectedDatasources) > 0 {
					return Result{
						Severity: Error,
						Message:  fmt.Sprintf("Dashboard '%s', panel '%s' does not use %s for datasource, uses '%s'", d.Title, p.Title, strings.Join(expectedDatasources, " or "), p.Datasource),
					}
				}
			}

			return ResultSuccess
		},
	}
}
