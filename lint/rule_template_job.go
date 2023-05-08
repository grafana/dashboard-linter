package lint

import (
	"fmt"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

	// Adding the prometheus_datasource here is hacky. This check function also assumes that all template vars which it will
	// ever check are only prometheus queries, which may not always be the case.
	if t.Datasource != "$datasource" && t.Datasource != "${datasource}" && t.Datasource != "$prometheus_datasource" && t.Datasource != "${prometheus_datasource}" {
		return &Result{
			Severity: Error,
			Message:  fmt.Sprintf("Dashboard '%s' %s template should use datasource '$datasource', is currently '%s'", d.Title, name, t.Datasource),
		}
	}

	if t.Type != targetTypeQuery {
		return &Result{
			Severity: Error,
			Message:  fmt.Sprintf("Dashboard '%s' %s template should be a Prometheus query, is currently '%s'", d.Title, name, t.Type),
		}
	}

	titleCaser := cases.Title(language.English)
	labelTitle := titleCaser.String(name)

	if t.Label != labelTitle {
		return &Result{
			Severity: Warning,
			Message:  fmt.Sprintf("Dashboard '%s' %s template should be a labeled '%s', is currently '%s'", d.Title, name, labelTitle, t.Label),
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
