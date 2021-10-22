package lint

import (
	"fmt"

	"github.com/grafana/cloud-onboarding/pkg/integrations-api/integrations"
)

func NewTemplateJobRule() *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "template-job-rule",
		description: "template-job-rule Checks that the dashboard has a templated job and instance.",
		fn: func(i *integrations.Integration, d Dashboard) Result {
			template := getTemplateDatasource(d)
			if template == nil || template.Query != "prometheus" {
				return Result{
					Severity: Success,
					Message:  "OK",
				}
			}

			{
				jobTemplate := getTemplate(d, "job")
				if jobTemplate == nil {
					return Result{
						Severity: Error,
						Message:  fmt.Sprintf("Dashboard '%s' is missing the job template", d.Title),
					}
				}

				if jobTemplate.Type != "prometheus" {
					return Result{
						Severity: Error,
						Message:  fmt.Sprintf("Dashboard '%s' job template should be a Prometheus query", d.Title),
					}
				}

				if jobTemplate.Label != "Job" {
					return Result{
						Severity: Error,
						Message:  fmt.Sprintf("Dashboard '%s' job template should be a labelled 'Job'", d.Title),
					}
				}
			}

			{
				instanceTemplate := getTemplate(d, "instance")
				if instanceTemplate == nil {
					return Result{
						Severity: Error,
						Message:  fmt.Sprintf("Dashboard '%s' is missing the instance template", d.Title),
					}
				}

				if instanceTemplate.Type != "prometheus" {
					return Result{
						Severity: Error,
						Message:  fmt.Sprintf("Dashboard '%s' instance template should be a Prometheus query", d.Title),
					}
				}

				if instanceTemplate.Label != "Instance" {
					return Result{
						Severity: Error,
						Message:  fmt.Sprintf("Dashboard '%s' instance template should be a labelled 'Instance'", d.Title),
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

func getTemplate(d Dashboard, name string) *Template {
	for _, template := range d.Templating.List {
		if template.Name == name {
			return &template
		}
	}
	return nil
}
