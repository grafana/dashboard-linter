package lint

import (
	"fmt"
)

func NewTemplateDatasourceRule() *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "template-datasource-rule",
		description: "Checks that the dashboard has a templated datasource.",
		fn: func(d Dashboard) Result {
			template := getTemplateDatasource(d)
			if template == nil {
				return Result{
					Severity: Error,
					Message:  fmt.Sprintf("Dashboard '%s' does not have a templated datasource", d.Title),
				}
			}

			if template.Name != "datasource" {
				return Result{
					Severity: Error,
					Message:  fmt.Sprintf("Dashboard '%s' templated datasource variable named '%s', should be names 'datasource'", d.Title, template.Name),
				}
			}

			if template.Label != "Data Source" {
				return Result{
					Severity: Error,
					Message:  fmt.Sprintf("Dashboard '%s' templated datasource variable labeled '%s', should be labeled 'Data Source'", d.Title, template.Label),
				}
			}

			if template.Query != "prometheus" && template.Query != "loki" {
				return Result{
					Severity: Error,
					Message:  fmt.Sprintf("Dashboard '%s' templated datasource variable query is '%s', should be 'prometheus' or 'loki'", d.Title, template.Query),
				}
			}

			return ResultSuccess
		},
	}
}

func getTemplateDatasource(d Dashboard) *Template {
	for _, template := range d.Templating.List {
		if template.Type != "datasource" {
			continue
		}
		return &template
	}
	return nil
}
