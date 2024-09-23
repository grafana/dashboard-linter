package lint

import (
	"fmt"
	"strings"
	"time"

	"github.com/grafana/loki/v3/pkg/logql/syntax"
)

func parseLogQL(expr string, variables []Template) (syntax.Expr, error) {
	expr, err := expandLogQLVariables(expr, variables)
	if err != nil {
		return nil, fmt.Errorf("could not expand variables: %w", err)
	}
	return syntax.ParseExpr(expr)
}

func NewTargetLogQLAutoRule() *TargetRuleFunc {
	autoDuration, err := time.ParseDuration(globalVariables["__auto"].(string))
	if err != nil {
		panic(err)
	}

	return &TargetRuleFunc{
		name:        "target-logql-auto-rule",
		description: "Checks that each Loki target uses $__auto for range vectors when appropriate.",
		fn: func(d Dashboard, p Panel, t Target) TargetRuleResults {
			r := TargetRuleResults{}

			// skip hidden targets
			if t.Hide {
				return r
			}

			// check if the datasource is Loki
			isLoki := false
			if templateDS := getTemplateDatasource(d); templateDS != nil && templateDS.Query == Loki {
				isLoki = true
			} else if ds, err := t.GetDataSource(); err == nil && ds.Type == Loki {
				isLoki = true
			}

			// skip if the datasource is not Loki
			if !isLoki {
				return r
			}

			// skip if the panel does not have queries
			if !panelHasQueries(p) {
				return r
			}

			parsedExpr, err := parseLogQL(t.Expr, d.Templating.List)
			if err != nil {
				r.AddError(d, p, t, fmt.Sprintf("Invalid LogQL query: %v", err))
				return r
			}

			originalExpr := t.Expr

			hasFixedDuration := false

			// Inspect the parsed expression to check for fixed durations
			Inspect(parsedExpr, func(node syntax.Expr) bool {
				if logRange, ok := node.(*syntax.LogRange); ok {
					if logRange.Interval != autoDuration && !strings.Contains(originalExpr, "$__auto") {
						hasFixedDuration = true
						return false
					}
				}
				return true
			})

			if hasFixedDuration {
				r.AddError(d, p, t, "LogQL query uses fixed duration: should use $__auto")
			}

			return r
		},
	}
}

func Inspect(node syntax.Expr, f func(syntax.Expr) bool) {
	if node == nil || !f(node) {
		return
	}
	switch n := node.(type) {
	case *syntax.BinOpExpr:
		Inspect(n.SampleExpr, f)
		Inspect(n.RHS, f)
	case *syntax.RangeAggregationExpr:
		Inspect(n.Left, f)
	case *syntax.VectorAggregationExpr:
		Inspect(n.Left, f)
	case *syntax.LabelReplaceExpr:
		Inspect(n.Left, f)
	case *syntax.LogRange:
		Inspect(n.Left, f)
	case *syntax.PipelineExpr:
		Inspect(n.Left, f)
		for _, stage := range n.MultiStages {
			f(stage)
		}
	}
}
