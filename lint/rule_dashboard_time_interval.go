package lint

import "fmt"

func NewDefaultTimeIntervalRule() *DashboardRuleFunc {
	timeIntervalFrom := "now-1h"
	timeIntervalTo := "now"

	return &DashboardRuleFunc{
		name:        "panel-time-interval-rule",
		description: "Checks that each panel has an appropriate time window interval.",
		fn: func(d Dashboard) Result {
			if d.Time.From != timeIntervalFrom || d.Time.To != timeIntervalTo {
				return Result{
					Severity: Error,
					Message:  fmt.Sprintf("Dashboard '%s' should have a default time interval of From: '%s' To: '%s', currently it is From: '%s' To: '%s'", d.Title, timeIntervalFrom, timeIntervalTo, d.Time.From, d.Time.To),
				}
			}
			return ResultSuccess
		},
	}
}
