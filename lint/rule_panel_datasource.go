package lint

import (
	"fmt"
)

func NewPanelDatasourceRule() *PanelRuleFunc {
	return &PanelRuleFunc{
		name:        "panel-datasource-rule",
		description: "Checks that each panel uses the templated datasource.",
		fn: func(d *Dashboard, p Panel) Result {
			switch p.Type {
			case panelTypeSingleStat, panelTypeGraph, panelTypeTimeTable, panelTypeTimeSeries:
				// That a templated datasource exists, is the responsibility of another rule.
				templatedDs := d.GetTemplateByType("datasource")
				availableDsUids := make(map[string]struct{}, len(templatedDs)*2)
				for _, tds := range templatedDs {
					availableDsUids[fmt.Sprintf("$%s", tds.Name)] = struct{}{}
					availableDsUids[fmt.Sprintf("${%s}", tds.Name)] = struct{}{}
				}

				src, err := p.GetDataSource()
				if err != nil {
					return Result{
						Severity: Error,
						Message:  fmt.Sprintf("Dashboard '%s', panel '%s' has invalid datasource: %v'", d.Title, p.Title, err),
					}
				}
				_, ok := availableDsUids[string(src)]
				if !ok {
					return Result{
						Severity: Error,
						Message:  fmt.Sprintf("Dashboard '%s', panel '%s' does not use a templated datasource, uses '%s'", d.Title, p.Title, src),
					}
				}
			}

			return ResultSuccess
		},
	}
}
