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
		stability:   ruleStabilityExperimental,
		fn: func(d Dashboard) DashboardRuleResults {
			r := DashboardRuleResults{}

			template := getTemplateDatasource(d)
			if template == nil || template.Query != Prometheus {
				return r
			}

			// Create a map and a slice, map for uniqueness and slice to keep the order...
			var varMap = make(map[string]bool)
			var varSlice = []string{}

			if config != nil {
				// Convert the config.variables to a map to leverage uniqueness...
				for _, v := range config.Variables {
					if varMap[v] {
						continue
					}
					varMap[v] = true
					varSlice = append(varSlice, v)
				}
			}

			if requiredMatchers != nil {
				// Check that all required matchers that use variables form target-required-matchers have a corresponding template variable
				for _, m := range requiredMatchers.Matchers {
					if strings.HasPrefix(m.Value, "$") {
						if varMap[m.Value[1:]] {
							continue
						}
						varMap[m.Value[1:]] = true
						varSlice = append(varSlice, m.Value[1:])
					}
				}
			}

			for _, v := range varSlice {
				checkTemplate(d, v, &r)
			}
			return r
		},
	}
}
