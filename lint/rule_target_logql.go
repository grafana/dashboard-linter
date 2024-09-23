package lint

import (
	"fmt"
)

func NewTargetLogQLRule() *TargetRuleFunc {
	return &TargetRuleFunc{
		name:        "target-logql-rule",
		description: "Checks that each target uses a valid LogQL query.",
		fn: func(d Dashboard, p Panel, t Target) TargetRuleResults {
			r := TargetRuleResults{}

			// Skip hidden targets
			if t.Hide {
				return r
			}

			// Check if the datasource is Loki
			isLoki := false
			if templateDS := getTemplateDatasource(d); templateDS != nil && templateDS.Query == Loki {
				isLoki = true
			} else if ds, err := t.GetDataSource(); err == nil && ds.Type == Loki {
				isLoki = true
			}

			// skip if the datasource is not Loki
			if !isLoki {
				return r
			}

			if !panelHasQueries(p) {
				return r
			}

			// If panel does not contain an expression then check if it references another panel and it exists
			if len(t.Expr) == 0 {
				if t.PanelId > 0 {
					for _, p1 := range d.Panels {
						if p1.Id == t.PanelId {
							return r
						}
					}
					r.AddError(d, p, t, "Invalid panel reference in target")
				}
				return r
			}

			// Parse the LogQL query
			_, err := parseLogQL(t.Expr, d.Templating.List)
			if err != nil {
				r.AddError(d, p, t, fmt.Sprintf("invalid LogQL query '%s': %v", t.Expr, err))
				return r
			}

			return r
		},
	}
}
