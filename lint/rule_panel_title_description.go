package lint

import "fmt"

func NewPanelTitleDescriptionRule() *PanelRuleFunc {
	return &PanelRuleFunc{
		name:        "panel-title-description-rule",
		description: "Checks that each panel has a title and description.",
		fn: func(d Dashboard, p Panel) Result {
			switch p.Type {
			case panelTypeStat, panelTypeSingleStat, panelTypeGraph, panelTypeTimeTable, panelTypeTimeSeries, panelTypeGauge:
				if len(p.Title) == 0 || len(p.Description) == 0 {
					return Result{
						Severity: Error,
						Message:  fmt.Sprintf("Dashboard '%s', panel with id '%v' has missing title or description, currently has title '%s' and description: '%s'", d.Title, p.Id, p.Title, p.Description),
					}
				}
			}
			return ResultSuccess
		},
	}
}
