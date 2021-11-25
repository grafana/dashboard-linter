package lint

import (
	"fmt"

	"github.com/prometheus/prometheus/promql/parser"
)

func NewTemplateLabelPromQLRule() *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "template-label-promql-rule",
		description: "Checks that the dashboard has a templated labels with proper promql expression.",
		fn: func(d Dashboard) Result {
			template := getTemplateDatasource(d)
			if template == nil || template.Query != "prometheus" {
				return Result{
					Severity: Success,
					Message:  "OK",
				}
			}
			for _, template := range d.Templating.List {
				if template.Type == "query" {
					if expr, err := parser.ParseExpr(template.Query); expr == nil {
						return Result{
							Severity: Error,
							Message:  fmt.Sprintf("Dashboard '%s', template '%s' invalid PromQL query '%s': %v", d.Title, template.Name, template.Query, err),
						}
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
