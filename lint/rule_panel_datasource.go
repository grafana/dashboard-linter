package lint

import (
	"fmt"
)

func NewPanelDatasourceRule() *PanelRuleFunc {
	return &PanelRuleFunc{
		name:        "panel-datasource-rule",
		description: "Checks that each panel uses the templated datasource.",
		fn: func(d Dashboard, p Panel) Result {
			switch p.Type {
			case "singlestat", "graph", "table", "timeseries":
				// That a templated datasource exists, is the responsibility of another rule.
				templatedDs := d.GetTemplateByType("datasource")
				availableDsUids := make(map[string]struct{}, len(templatedDs)*2)
				for _, tds := range templatedDs {
					availableDsUids[fmt.Sprintf("$%s", tds.Name)] = struct{}{}
					availableDsUids[fmt.Sprintf("${%s}", tds.Name)] = struct{}{}
				}

				_, ok := availableDsUids[string(p.Datasource)]
				if !ok {
					return Result{
						Severity: Error,
						Message:  fmt.Sprintf("Dashboard '%s', panel '%s' does not use a templated datasource, uses '%s'", d.Title, p.Title, p.Datasource),
					}
				}
			}

			return ResultSuccess
		},
	}
}
