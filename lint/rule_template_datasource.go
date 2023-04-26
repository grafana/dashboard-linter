package lint

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func NewTemplateDatasourceRule() *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "template-datasource-rule",
		description: "Checks that the dashboard has a templated datasource.",
		fn: func(d Dashboard) Result {
			templatedDs := d.GetTemplateByType("datasource")
			if len(templatedDs) == 0 {
				return Result{
					Severity: Error,
					Message:  fmt.Sprintf("Dashboard '%s' does not have a templated data source", d.Title),
				}
			}

			// TODO: Should there be a "Template" rule type which will iterate over all dashboard templates and execute rules?
			// This will only return one linting error at a time, when there may be multiple issues with templated datasources.

			titleCaser := cases.Title(language.English)

			for _, templDs := range templatedDs {
				querySpecificUID := fmt.Sprintf("%s_datasource", strings.ToLower(templDs.Query))
				querySpecificName := fmt.Sprintf("%s Data Source", titleCaser.String(templDs.Query))

				allowedDsUIDs := make(map[string]struct{})
				allowedDsNames := make(map[string]struct{})

				uidError := fmt.Sprintf("Dashboard '%s' templated data source variable named '%s', should be named '%s'", d.Title, templDs.Name, querySpecificUID)
				nameError := fmt.Sprintf("Dashboard '%s' templated data source variable labeled '%s', should be labeled '%s'", d.Title, templDs.Label, querySpecificName)
				if len(templatedDs) == 1 {
					allowedDsUIDs["datasource"] = struct{}{}
					allowedDsNames["Data Source"] = struct{}{}

					uidError = uidError + ", or 'datasource'"
					nameError = nameError + ", or 'Data Source'"
				}

				allowedDsUIDs[querySpecificUID] = struct{}{}
				allowedDsNames[querySpecificName] = struct{}{}

				// TODO: These are really two different rules
				_, ok := allowedDsUIDs[templDs.Name]
				if !ok {
					return Result{
						Severity: Error,
						Message:  uidError,
					}
				}

				_, ok = allowedDsNames[templDs.Label]
				if !ok {
					return Result{
						Severity: Error,
						Message:  nameError,
					}
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
