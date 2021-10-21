package lint

import (
	"fmt"

	"github.com/grafana/cloud-onboarding/pkg/integrations-api/integrations"
)

// NewPanelRateIntervalRule builds a lint rule for panels with Prometheus queries which checks
// all range vector selectors use $__rate_interval.
func NewPanelRateIntervalRule() *PanelRuleFunc {
	return &PanelRuleFunc{
		name:        "panel-rate-interval-rule",
		description: "panel-rate-interval-rule Checks that each panel uses $__rate_interval.",
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
					for _, rate := range rangeVectorRegexp.FindAllString(target.Expr, -1) {
						if rate != "[$__rate_interval]" {
							return Result{
								Severity: Error,
								Message:  fmt.Sprintf("Dashboard '%s', panel '%s' invalid PromQL query '%s': should use $__rate_interval", d.Title, p.Title, target.Expr),
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
