package lint

import "fmt"

func NewDefaultTimeRefresh() *DashboardRuleFunc {
	timeIntervalFrom := "now-1h"
	timeIntervalTo := "now"
	refreshInterval := "5m"

	return &DashboardRuleFunc{
		name:        "panel-time-refresh-rule",
		description: "Checks that each panel has an appropriate time window and refresh interval.",
		fn: func(d Dashboard) Result {
			if d.Time.From != timeIntervalFrom && d.Time.To != timeIntervalTo {
				return Result{
					Severity: Error,
					Message:  fmt.Sprintf("Dashboard '%s' should have a default time interval of last 1 hour, current it is From: '%s' To: '%s'", d.Title, d.Time.From, d.Time.To),
				}
			}
			if d.Refresh != refreshInterval {
				return Result{
					Severity: Error,
					Message:  fmt.Sprintf("Dashboard '%s' should have a default refresh interval of last 1 hour, current it is From: '%s' To: '%s'", d.Title, d.Time.From, d.Time.To),
				}
			}
			return ResultSuccess
		},
	}
}
