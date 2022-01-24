package lint

import (
	"fmt"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// NewPanelJobInstanceRule builds a lint rule for panels with Prometheus queries which checks
// the query contains two matchers within every selector - `{job=~"$job", instance=~"$instance"}`
func NewPanelJobInstanceRule() *PanelRuleFunc {
	return &PanelRuleFunc{
		name:        "panel-job-instance-rule",
		description: "Checks that every PromQL query has job and instance matchers.",
		fn: func(d Dashboard, p Panel) Result {
			if t := getTemplateDatasource(d); t == nil || t.Query != "prometheus" {
				// Missing template datasources is a separate rule.
				return Result{
					Severity: Success,
					Message:  "OK",
				}
			}

			if !panelHasQueries(p) {
				return Result{
					Severity: Success,
					Message:  "OK",
				}
			}

			for _, target := range p.Targets {
				node, err := parsePromQL(target.Expr)
				if err != nil {
					// Invalid PromQL is another rule.
					return Result{
						Severity: Success,
						Message:  "OK",
					}
				}

				for _, selector := range parser.ExtractSelectors(node) {
					if err := checkForMatcher(selector, "job", labels.MatchRegexp, "$job"); err != nil {
						return Result{
							Severity: Error,
							Message:  fmt.Sprintf("Dashboard '%s', panel '%s' invalid PromQL query '%s': %v", d.Title, p.Title, target.Expr, err),
						}
					}

					if err := checkForMatcher(selector, "instance", labels.MatchRegexp, "$instance"); err != nil {
						return Result{
							Severity: Error,
							Message:  fmt.Sprintf("Dashboard '%s', panel '%s' invalid PromQL query '%s': %v", d.Title, p.Title, target.Expr, err),
						}
					}
				}
			}

			return Result{
				Severity: Success,
				Message:  "OK",
			}
		},
	}
}

func checkForMatcher(selector []*labels.Matcher, name string, ty labels.MatchType, value string) error {
	for _, matcher := range selector {
		if matcher.Name != name {
			continue
		}

		if matcher.Type != ty {
			return fmt.Errorf("%s selector is %s, not %s", name, matcher.Type, ty)
		}

		if matcher.Value != value {
			return fmt.Errorf("%s selector is %s, not %s", name, matcher.Value, value)
		}

		return nil
	}

	return fmt.Errorf("%s selector not found", name)
}
