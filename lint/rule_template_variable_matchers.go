package lint

import (
	"strings"

	"github.com/prometheus/prometheus/model/labels"
)

func NewTemplateVariableMatchersRule(matchers []*labels.Matcher) *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "template-variable-matchers-rule",
		description: "Checks that the dashboard has a template variable for required matchers that use variables",
		fn: func(d Dashboard) DashboardRuleResults {
			r := DashboardRuleResults{}

			template := getTemplateDatasource(d)
			if template == nil || template.Query != Prometheus {
				return r
			}

			for _, m := range matchers {
				if strings.HasPrefix(m.Value, "$") {
					checkTemplate(d, m.Value[1:], &r)
				}
			}
			return r
		},
	}
}
