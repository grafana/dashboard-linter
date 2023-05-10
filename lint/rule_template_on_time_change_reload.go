package lint

import (
	"fmt"
)

func NewTemplateOnTimeRangeReloadRule() *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "template-on-time-change-reload-rule",
		description: "Checks that the dashboard template variables are configured to reload on time change.",
		fn: func(d *Dashboard, cfg *ConfigurationFile) Result {
			var l = []Template{}
			for i, template := range d.Templating.List {
				l = append(l, template)
				if template.Type != targetTypeQuery {
					continue
				}

				if template.Refresh != 2 {
					if cfg.Autofix {
						template.Refresh = 2
						l[i] = template
					} else {
						return Result{
							Severity: Error,
							Message:  fmt.Sprintf("Dashboard '%s' templated datasource variable named '%s', should be set to be refreshed 'On Time Range Change (value 2)', is currently '%d'", d.Title, template.Name, template.Refresh),
						}
					}
				}
			}
			d.Templating.List = l

			return Result{
				Severity: Success,
				Message:  "OK",
			}
		},
	}
}
