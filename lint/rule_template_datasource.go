package lint

import (
	"fmt"
	"sort"
	"strings"
)

// This rule checks that the dashboard has single templated datasource
func (d *Dashboard) checkPrometheusOrLokiDS(template *Template) Result {

	if template.Name != "datasource" {
		return Result{
			Severity: Error,
			Message:  fmt.Sprintf("Dashboard '%s' templated datasource variable named '%s', should be named 'datasource'", d.Title, template.Name),
		}
	}

	if template.Label != "Data Source" {
		return Result{
			Severity: Error,
			Message:  fmt.Sprintf("Dashboard '%s' templated datasource variable labeled '%s', should be labeled 'Data Source'", d.Title, template.Label),
		}
	}

	if template.Query != Prometheus && template.Query != Loki {
		return Result{
			Severity: Error,
			Message:  fmt.Sprintf("Dashboard '%s' templated datasource variable query is '%s', should be 'prometheus' or 'loki'", d.Title, template.Query),
		}
	}
	return ResultSuccess
}

func (d *Dashboard) checkPrometheusAndLokiDS(datasources []Template) Result {

	// move loki to 0 index
	sort.SliceStable(datasources, func(i, j int) bool {
		return datasources[i].Query < datasources[j].Query
	})

	if !(datasources[0].Query == Loki && datasources[1].Query == Prometheus) {
		return Result{
			Severity: Error,
			Message:  fmt.Sprintf("Dashboard '%s' with 2 templated datasources should have 'prometheus' and 'loki' types", d.Title),
		}
	}

	for _, template := range datasources {
		prefix := template.Query

		expected_name := prefix + "_datasource"
		if template.Name != expected_name {
			return Result{
				Severity: Error,
				Message:  fmt.Sprintf("Dashboard '%s' templated datasource variable named '%s', should be named '%s'", d.Title, template.Name, expected_name),
			}
		}

		expected_label := strings.Title(prefix) + " Data Source"
		if template.Label != expected_label {
			return Result{
				Severity: Error,
				Message:  fmt.Sprintf("Dashboard '%s' templated datasource variable labeled '%s', should be labeled '%s'", d.Title, template.Label, expected_label),
			}
		}
	}

	return ResultSuccess
}

func NewTemplateDatasourceRule() *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "template-datasource-rule",
		description: "Checks that the dashboard has a templated datasource.",
		fn: func(d Dashboard) Result {

			datasources := getTemplateDatasources(d)
			if len(datasources) == 1 {
				return d.checkPrometheusOrLokiDS(&datasources[0])
			} else if len(datasources) == 2 {
				return d.checkPrometheusAndLokiDS(datasources)
			} else {
				return Result{
					Severity: Error,
					Message:  fmt.Sprintf("Dashboard '%s' has %d templated datasources, should be 1 or 2", d.Title, len(datasources)),
				}
			}
		},
	}
}

//Returns only first datasource found
func getTemplateDatasource(d Dashboard) *Template {
	for _, template := range d.Templating.List {
		if template.Type != "datasource" {
			continue
		}
		return &template
	}
	return nil
}

func getTemplateDatasources(d Dashboard) []Template {

	templateDatasources := []Template{}
	for _, template := range d.Templating.List {
		if template.Type != "datasource" {
			continue
		}
		templateDatasources = append(templateDatasources, template)
	}
	return templateDatasources
}

// Returns datasource names that should be used instead of provided datasource
func checkTemplatedDatasourceUsed(d Dashboard, datasource Datasource) []string {

	datasource_count := len(getTemplateDatasources(d))
	if datasource_count == 1 {
		if datasource != "$datasource" && datasource != "${datasource}" {
			return []string{"$datasource"}
		}
	} else if datasource_count == 2 {

		if datasource != "$prometheus_datasource" && datasource != "${prometheus_datasource}" &&
			datasource != "$loki_datasource" && datasource != "${loki_datasource}" {
			return []string{"$prometheus_datasource", "$loki_datasource"}
		}
	}
	return []string{}
}
