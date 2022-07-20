package lint

import (
	"fmt"
	"strings"

	"github.com/prometheus/prometheus/promql/parser"
)

func NewTargetHistogramRule() *TargetRuleFunc {
	return &TargetRuleFunc{
		name:        "target-histogram-rule",
		description: "Checks that any bucket metric (ending in _bucket) is calculated with histogram.",
		fn: func(d Dashboard, p Panel, t Target) Result {
			expr, err := parsePromQL(t.Expr, d.Templating.List)
			if err != nil {
				// Invalid PromQL is another rule
				return ResultSuccess
			}

			err = parser.Walk(inspector(func(node parser.Node, parents []parser.Node) error {
				// We're looking for either a VectorSelector. This skips any other node type.
				selector, ok := node.(*parser.VectorSelector)
				if !ok {
					return nil
				}

				errmsg := fmt.Errorf("Dashboard '%s', panel '%s', target idx '%d' histogram metric '%s' is not calculated in a histogram function", d.Title, p.Title, t.Idx, node.String())

				if strings.HasSuffix(selector.String(), "_bucket") {
					// The vector selector must have (at least) one parent
					if len(parents) == 0 {
						return errmsg
					}

					// Just reverse walk the stack of parents and return success if we find a histogram_quantile call
					currentParent := len(parents) - 1
					for currentParent >= 0 {
						call, ok := parents[currentParent].(*parser.Call)
						currentParent -= 1
						if !ok {
							continue
						}
						if call.Func.Name == "histogram_quantile" {
							return nil
						}
					}
					return errmsg
				}
				return nil
			}), expr, nil)
			if err != nil {
				return Result{
					Severity: Error,
					Message:  err.Error(),
				}
			}
			return ResultSuccess
		},
	}
}
