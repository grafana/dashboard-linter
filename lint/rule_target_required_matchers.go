package lint

import (
	"fmt"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

func newTargetRequiredMatchersRule(matchers []*labels.Matcher) *TargetRuleFunc {
	return &TargetRuleFunc{
		name:        "target-required-matchers",
		description: "Checks that target expr has the required matchers",
		fn: func(d Dashboard, p Panel, t Target) TargetRuleResults {
			r := TargetRuleResults{}
			// TODO: The RuleSet should be responsible for routing rule checks based on their query type (prometheus, loki, mysql, etc)
			// and for ensuring that the datasource is set.
			if t := getTemplateDatasource(d); t == nil || t.Query != Prometheus {
				// Missing template datasource is a separate rule.
				// Non prometheus datasources don't have rules yet
				return r
			}

			expr, err := parsePromQL(t.Expr, d.Templating.List)
			if err != nil {
				// Invalid PromQL is another rule
				return r
			}
			for _, m := range matchers {
				for _, selector := range parser.ExtractSelectors(expr) {
					if err := checkForMatcher(selector, m.Name, m.Type, m.Value); err != nil {
						r.AddFixableError(d, p, t, fmt.Sprintf("invalid PromQL query '%s': %v", t.Expr, err), fixTargetRequiredMatcherRule(m.Name, m.Type, m.Value))
					}
				}
			}
			return r
		},
	}
}

func fixTargetRequiredMatcherRule(name string, ty labels.MatchType, value string) func(Dashboard, Panel, *Target) {
	return func(d Dashboard, p Panel, t *Target) {
		// using t.Expr to ensure matchers added earlier in the loop are not lost
		// no need to check for errors here, as the expression was already parsed and validated
		expr, _ := parsePromQL(t.Expr, d.Templating.List)
		// Walk the expression tree and add the matcher to all vector selectors
		parser.Walk(addMatchers(name, ty, value), expr, nil)
		t.Expr = expr.String()
	}
}

type matcherAdder func(node parser.Node) error

func (f matcherAdder) Visit(node parser.Node, path []parser.Node) (w parser.Visitor, err error) {
	err = f(node)
	return f, err
}

func addMatchers(name string, ty labels.MatchType, value string) matcherAdder {
	return func(node parser.Node) error {
		if n, ok := node.(*parser.VectorSelector); ok {
			matcherfixed := false
			for _, m := range n.LabelMatchers {
				if m.Name == name {
					if m.Type != ty || m.Value != value {
						m.Type = ty
						m.Value = value
					}
					matcherfixed = true
				}
			}
			if !matcherfixed {
				n.LabelMatchers = append(n.LabelMatchers, &labels.Matcher{
					Name:  name,
					Type:  ty,
					Value: value,
				})
			}
		}
		return nil
	}
}