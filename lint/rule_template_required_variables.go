package lint

import (
	"strings"
)

type TemplateRequiredVariablesRuleSettings struct {
	Variables []string `yaml:"variables"`
}

func NewTemplateRequiredVariablesRule(config *TemplateRequiredVariablesRuleSettings, requiredMatchers *TargetRequiredMatchersRuleSettings) *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "template-required-variables-rule",
		description: "Checks that the dashboard has a template variable for required variables or matchers that use variables",
		stability:   "experimental",
		fn: func(d Dashboard) DashboardRuleResults {
			r := DashboardRuleResults{}

			template := getTemplateDatasource(d)
			if template == nil || template.Query != Prometheus {
				return r
			}

			// Convert the config.variables to a map to leverage uniqueness...
			variables := make(map[string]bool)
			for _, v := range config.Variables {
				variables[v] = true
			}

			// Check that all required matchers that use variables form target-required-matchers have a corresponding template variable
			for _, m := range requiredMatchers.Matchers {
				if strings.HasPrefix(m.Value, "$") {
					variables[m.Value[1:]] = true
				}
			}

			for v := range variables {
				checkTemplate(d, v, &r)
			}
			return r
		},
	}
}
