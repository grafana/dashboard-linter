package lint

import (
	"fmt"

	"github.com/prometheus/prometheus/promql/parser"
)

// panelHasQueries returns true is the panel has queries we should try and
// validate.  We allow-list panels here to prevent false positives with
// new panel types we don't understand.
func panelHasQueries(p Panel) bool {
	types := []string{"singlestat", "graph", "table", "stat", "state-timeline", "timeseries"}
	for _, t := range types {
		if p.Type == t {
			return true
		}
	}
	return false
}

// parsePromQL returns the parsed PromQL statement from a panel,
// replacing eg [$__rate_interval] with [5m] so queries parse correctly.
// We also replace various other Grafana global variables.
func parsePromQL(expr string, variables []Template) (parser.Expr, error) {
	expr, err := expandVariables(expr, variables)
	if err != nil {
		return nil, fmt.Errorf("could not expand variables: %w", err)
	}
	return parser.ParseExpr(expr)
}

// NewPanelPromQLRule builds a lint rule for panels with Prometheus queries which checks:
// - the query is valid PromQL
// - the query contains two matchers within every selector - `{job=~"$job", instance=~"$instance"}`
func NewPanelPromQLRule() *PanelRuleFunc {
	return &PanelRuleFunc{
		name:        "panel-promql-rule",
		description: "Checks that each panel uses a valid PromQL query.",
		fn: func(d Dashboard, p Panel) Result {
			if t := getTemplateDatasource(d); t == nil || t.Query != "prometheus" {
				// Missing template datasources is a separate rule.
				return ResultSuccess
			}

			if !panelHasQueries(p) {
				return ResultSuccess
			}

			for _, target := range p.Targets {
				if _, err := parsePromQL(target.Expr, d.Templating.List); err != nil {
					return Result{
						Severity: Error,
						Message:  fmt.Sprintf("Dashboard '%s', panel '%s' invalid PromQL query '%s': %v", d.Title, p.Title, target.Expr, err),
					}
				}
			}

			return ResultSuccess
		},
	}
}
