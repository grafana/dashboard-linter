package lint

import (
	"fmt"
	"regexp"
	"strings"
)

func targetHasTemplate(target Target, varRegexp *regexp.Regexp) bool {
	return varRegexp.FindStringSubmatch(target.Expr) != nil
}

func NewTemplateUnusedRule() *TemplateRuleFunc {
	return &TemplateRuleFunc{
		name:        "template-unused-rule",
		description: "Checks that each query template is used in queries.",
		fn: func(d Dashboard, t Template) Result {

			// check only query templates for now
			if t.Type != "query" {
				return ResultSuccess
			}

			var variableRegexp = regexp.MustCompile(
				strings.Join([]string{
					`\$(` + t.Name + `)\b`,            // $var syntax
					`\$\{(` + t.Name + `(:[^}]+)?)\}`, // ${var} syntax including ${var_name:<format>}
					`\[\[(` + t.Name + `)\]\]`,        // [[var]] deprecetaed syntax
				}, "|"),
			)

			for _, p := range d.Panels {
				for _, target := range p.Targets {
					if targetHasTemplate(target, variableRegexp) {
						return ResultSuccess
					}
				}

			}
			return Result{
				Severity: Error,
				Message:  fmt.Sprintf("Dashboard '%s', template '%s' is not used in any panel.", d.Title, t.Name),
			}
		},
	}
}
