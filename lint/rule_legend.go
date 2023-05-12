package lint

import (
	"fmt"
	"strings"
)

const legendPrefix = "{{instance}} - "

func NewLegendRule() *TargetRuleFunc {
	matcher := "foo"
	return &TargetRuleFunc{
		name:        fmt.Sprintf("target-%s-rule", matcher),
		description: fmt.Sprintf("Checks that every PromQL query has a %s matcher.", matcher),
		fn: func(d Dashboard, p Panel, t Target) Result {
			switch p.Type {
			case "stat", "singlestat", "graph", "table", "timeseries", "gauge", "barchart", "bargauge", "piechart", "histogram":
				break
			default:
				return ResultSuccess
			}

			l := t.LegendFormat
			if l == "" {
				return NewErrorResult(d, p, t, fmt.Sprintf("Legend is missing - should start with '%s'", legendPrefix))
			}
			if strings.HasPrefix(l, legendPrefix) {
				return ResultSuccess
			} else {
				return NewErrorResult(d, p, t, fmt.Sprintf("Legend should start with '%s' - found '%s'", legendPrefix, l))
			}
		},
	}
}
