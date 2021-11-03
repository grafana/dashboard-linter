package lint

import (
	"fmt"
)

func NewTemplateJobRule() *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "template-job-rule",
		description: "template-job-rule Checks that the dashboard has a templated job and instance.",
		fn: func(d Dashboard) Result {
			template := getTemplateDatasource(d)
			if template == nil || template.Query != "prometheus" {
				return Result{
					Severity: Success,
					Message:  "OK",
				}
			}

			if r := checkTemplate(d, "job"); r != nil {
				return *r
			}

			if r := checkTemplate(d, "instance"); r != nil {
				return *r
			}

			return Result{
				Severity: Success,
				Message:  "OK",
			}
		},
	}
}

func checkTemplate(d Dashboard, name string) *Result {
	t := getTemplate(d, name)
	if t == nil {
		return &Result{
			Severity: Error,
			Message:  fmt.Sprintf("Dashboard '%s' is missing the %s template", d.Title, name),
		}
	}

	if t.Datasource != "$datasource" {
		return &Result{
			Severity: Error,
			Message:  fmt.Sprintf("Dashboard '%s' %s template should use datasource '$datasource'", d.Title, name),
		}
	}

	if t.Type != "query" {
		return &Result{
			Severity: Error,
			Message:  fmt.Sprintf("Dashboard '%s' %s template should be a Prometheus query", d.Title, name),
		}
	}

	if t.Label != name {
		return &Result{
			Severity: Error,
			Message:  fmt.Sprintf("Dashboard '%s' %s template should be a labelled '%s'", d.Title, name, name),
		}
	}

	if !t.Multi {
		return &Result{
			Severity: Error,
			Message:  fmt.Sprintf("Dashboard '%s' %s template should be a multi select", d.Title, name),
		}
	}

	if t.AllValue != ".+" {
		return &Result{
			Severity: Error,
			Message:  fmt.Sprintf("Dashboard '%s' %s template allValue should be '.+'", d.Title, name),
		}
	}

	return &Result{
		Severity: Success,
		Message:  "OK",
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
