package lint

import (
	"fmt"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

func NewTargetJobRule() *TargetRuleFunc {
	return &TargetRuleFunc{
		name:        "target-job-rule",
		description: "Checks that every PromQL query has a job matcher.",
		fn: func(d Dashboard, p Panel, t Target) Result {
			// TODO: The RuleSet should be responsible for routing rule checks based on their query type (prometheus, loki, mysql, etc)
			// and for ensuring that the datasource is set.
			if t := getTemplateDatasource(d); t == nil || t.Query != Prometheus {
				// Missing template datasource is a separate rule.
				// Non prometheus datasources don't have rules yet
				return ResultSuccess
			}

			node, err := parsePromQL(t.Expr, d.Templating.List)
			if err != nil {
				// Invalid PromQL is another rule
				return ResultSuccess
			}

			for _, selector := range parser.ExtractSelectors(node) {
				if err := checkForMatcher(selector, "job", labels.MatchRegexp, "$job"); err != nil {
					return Result{
						Severity: Error,
						Message:  fmt.Sprintf("Dashboard '%s', panel '%s', target idx '%d' invalid PromQL query '%s': %v", d.Title, p.Title, t.Idx, t.Expr, err),
					}
				}
			}

			return ResultSuccess
		},
	}
}
