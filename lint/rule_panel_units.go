package lint

func NewPanelUnitsRule() *PanelRuleFunc {
	return &PanelRuleFunc{
		name:        "panel-units-rule",
		description: "Checks that each panel uses has units defined.",
		fn: func(d Dashboard, p Panel) Result {
			switch p.Type {
			case "singlestat", "graph", "table", "timeseries":

				// if p.Datasource != "$datasource" && p.Datasource != "${datasource}" {
				// 	return Result{
				// 		Severity: Error,
				// 		Message:  fmt.Sprintf("Dashboard '%s', panel '%s' does not use templates datasource, uses '%s'", d.Title, p.Title, p.Datasource),
				// 	}
				// }
			}

			return ResultSuccess
		},
	}
}
