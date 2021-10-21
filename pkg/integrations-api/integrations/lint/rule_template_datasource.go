package lint

import (
	"fmt"

	"github.com/grafana/cloud-onboarding/pkg/integrations-api/integrations"
)

func NewTemplateDatasourceRule() *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "template-datasource-rule",
		description: "template-datasource-rule Checks that the dashboard has a templated datasource.",
		fn: func(i *integrations.Integration, d Dashboard) Result {
			for _, template := range d.Templating.List {
				if template.Type != "datasource" {
					continue
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

				return Result{
					Severity: Success,
					Message:  "OK",
				}
			}

			return Result{
				Severity: Error,
				Message:  fmt.Sprintf("Dashboard '%s' does not have a templated datasource", d.Title),
			}
		},
	}
}
