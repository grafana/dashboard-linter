package lint

import (
	"fmt"
)

func NewTemplateJobRule() *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "template-job-rule",
		description: "Checks that the dashboard has a templated job.",
		fn: func(d Dashboard) Result {
			template := getTemplateDatasource(d)
			if template == nil || template.Query != Prometheus {
				return ResultSuccess
			}

			if r := checkTemplate(d, "job"); r != nil {
				return *r
			}

			return ResultSuccess
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

	if t.Datasource != "$datasource" && t.Datasource != "${datasource}" {
		return &Result{
			Severity: Error,
			Message:  fmt.Sprintf("Dashboard '%s' %s template should use datasource '$datasource', is currently '%s'", d.Title, name, t.Datasource),
		}
	}

	if t.Type != "query" {
		return &Result{
			Severity: Error,
			Message:  fmt.Sprintf("Dashboard '%s' %s template should be a Prometheus query, is currently '%s'", d.Title, name, t.Type),
		}
	}

	if t.Label != name {
		return &Result{
			Severity: Error,
			Message:  fmt.Sprintf("Dashboard '%s' %s template should be a labelled '%s', is currently '%s'", d.Title, name, name, t.Label),
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
			Message:  fmt.Sprintf("Dashboard '%s' %s template allValue should be '.+', is currently '%s'", d.Title, name, t.AllValue),
		}
	}

	return nil
}

func getTemplate(d Dashboard, name string) *Template {
	for _, template := range d.Templating.List {
		if template.Name == name {
			return &template
		}
	}
	return nil
}
