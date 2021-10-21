package lint

import (
	"fmt"
	"regexp"

	"github.com/grafana/cloud-onboarding/pkg/integrations-api/integrations"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

var rangeVectorRegexp = regexp.MustCompile(`\[\$\w+\]`)
var subqueryRegexp = regexp.MustCompile(`\[\$\w+:(.+)?\]`)

// NewPanelPromQLRule builds a lint rule for panels with Prometheus queries which checks:
// - the query is valid PromQL
// - the query contains two matchers within every selector - `{job=~"$job", instance=~"$instance"}`
func NewPanelPromQLRule() *PanelRuleFunc {
	return &PanelRuleFunc{
		name:        "panel-promql-rule",
		description: "panel-promql-rule Checks that each panel uses a valid PromQL query.",
		fn: func(i *integrations.Integration, d Dashboard, p Panel) Result {
			if t := getTemplateDatasource(d); t == nil || t.Query != "prometheus" {
				// Missing template datasources is a separate rule.
				return Result{
					Severity: Success,
					Message:  "OK",
				}
			}

			switch p.Type {
			case "singlestat", "graph", "table":
				for _, target := range p.Targets {
					// Hack in replace [$__rate_interval] with [5m] so queries parse correctly.
					expr := rangeVectorRegexp.ReplaceAllString(target.Expr, "[5m]")
					expr = subqueryRegexp.ReplaceAllString(expr, "[5m:]")
					node, err := parser.ParseExpr(expr)
					if err != nil {
						return Result{
							Severity: Error,
							Message:  fmt.Sprintf("Dashboard '%s', panel '%s' invalid PromQL query '%s': %v", d.Title, p.Title, target.Expr, err),
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
