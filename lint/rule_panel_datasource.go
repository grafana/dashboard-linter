package lint

import (
	"fmt"
)

// builtinDatasourceType is the plugin type of the datasources that Grafana
// provides itself: the "-- Grafana --", "-- Mixed --" and "-- Dashboard --"
// entries of the datasource picker.
const builtinDatasourceType = "datasource"

// grafanaDatasourceUID is the fixed UID of the "-- Grafana --" datasource.
// Grafana resolves any other explicit UID without looking at the type, so the
// type alone does not make a datasource built in.
const grafanaDatasourceUID = "grafana"

// builtinDatasourceUIDs are the names that older dashboards use for the
// built-in datasources, where the datasource is a plain string. A datasource
// variable cannot stand in for any of them, so a panel that uses one is not
// a portability problem.
var builtinDatasourceUIDs = map[string]struct{}{
	"-- Grafana --":   {},
	"-- Mixed --":     {},
	"-- Dashboard --": {},
}

func isBuiltinDatasource(ds Datasource) bool {
	if ds.Type == builtinDatasourceType && ds.UID == grafanaDatasourceUID {
		return true
	}
	_, ok := builtinDatasourceUIDs[ds.UID]
	return ok
}

func NewPanelDatasourceRule() *PanelRuleFunc {
	return &PanelRuleFunc{
		name:        "panel-datasource-rule",
		description: "Checks that each panel uses the templated datasource.",
		fn: func(d Dashboard, p Panel) PanelRuleResults {
			r := PanelRuleResults{}

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
					r.AddError(d, p, fmt.Sprintf("has invalid datasource: %v'", err))
				}
				if isBuiltinDatasource(src) {
					return r
				}
				_, ok := availableDsUids[string(src.UID)]
				if !ok {
					r.AddError(d, p, fmt.Sprintf("does not use a templated datasource, uses '%s'", src.UID))
				}
			}

			return r
		},
	}
}
