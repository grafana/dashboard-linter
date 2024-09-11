package lint

import (
	"fmt"
	"regexp"
)

var (
	lvNoQueryRegexp = regexp.MustCompile(`(?s)label_values\((.+)\)`)    // label_values(label)
	lvRegexp        = regexp.MustCompile(`(?s)label_values\((.+),.+\)`) // label_values(metric, label)
	mRegexp         = regexp.MustCompile(`(?s)metrics\((.+)\)`)         // metrics(metric)
	lnRegexp        = regexp.MustCompile(`(?s)label_names\((.+)\)`)     // label_names()
	qrRegexp        = regexp.MustCompile(`(?s)query_result\((.+)\)`)    // query_result(query)
)

func extractPromQLQuery(q string) []string {
	// label_values(query, label)
	switch {
	case lvRegexp.MatchString(q):
		return lvRegexp.FindStringSubmatch(q)
	case lvNoQueryRegexp.MatchString(q):
		return nil // No query so no metric.
	case mRegexp.MatchString(q):
		return mRegexp.FindStringSubmatch(q)
	case lnRegexp.MatchString(q):
		return lnRegexp.FindStringSubmatch(q)
	case qrRegexp.MatchString(q):
		return qrRegexp.FindStringSubmatch(q)
	default:
		return nil
	}
}

// parseTemplatedLabelPromQL returns error in case
// 1) The given PromQL expressions is invalid
// 2) Use of invalid label function
func parseTemplatedLabelPromQL(t Template, variables []Template) error {
	// regex capture must return slice of 2 strings.
	// 1) given query 2) function arg.

	tokens := extractPromQLQuery(t.Query)
	if tokens == nil {
		return fmt.Errorf("invalid 'query': %v", t.Query)
	}

	expr, err := parsePromQL(tokens[1], variables)
	if expr != nil {
		return nil
	}
	return err
}

func NewTemplateLabelPromQLRule() *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "template-label-promql-rule",
		description: "Checks that the dashboard templated labels have proper PromQL expressions.",
		stability:   ruleStabilityStable,
		fn: func(d Dashboard) DashboardRuleResults {
			r := DashboardRuleResults{}

			template := getTemplateDatasource(d)
			if template == nil || template.Query != Prometheus {
				return r
			}
			for _, template := range d.Templating.List {
				if template.Type != targetTypeQuery {
					continue
				}
				if err := parseTemplatedLabelPromQL(template, d.Templating.List); err != nil {
					r.AddError(d, fmt.Sprintf("template '%s' invalid templated label '%s': %v", template.Name, template.Query, err))
				}
			}
			return r
		},
	}
}
