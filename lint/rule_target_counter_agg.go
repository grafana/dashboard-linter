package lint

import (
	"errors"
	"fmt"
	"strings"

	"github.com/prometheus/prometheus/promql/parser"
)

func NewTargetCounterAggRule() *TargetRuleFunc {
	return &TargetRuleFunc{
		name:        "target-counter-agg-rule",
		description: "Checks that any counter metric (ending in _total) is aggregated with rate, irate, or increase.",
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

				errmsg := errors.New(NewErrorMessage(d, p, t,
					fmt.Sprintf("counter metric '%s' is not aggregated with rate, irate, or increase", node.String())))

				if strings.HasSuffix(selector.String(), "_total") {
					// The vector selector must have (at least) two parents
					if len(parents) < 2 {
						return errmsg
					}
					// The vector must be ranged
					_, ok := parents[len(parents)-1].(*parser.MatrixSelector)
					if !ok {
						return errmsg
					}
					// The range, must be in a function call
					call, ok := parents[len(parents)-2].(*parser.Call)
					if !ok {
						return errmsg
					}
					// Finally, the immediate ancestor call must be rate, irate, or increase
					if call.Func.Name != "rate" && call.Func.Name != "irate" && call.Func.Name != "increase" {
						return errmsg
					}
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
